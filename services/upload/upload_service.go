package upload

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ambroise1219/livraison_go/db"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/services/storage"
	"github.com/sirupsen/logrus"
)

// Service gère la logique d'upload d'images/documents
// Il s'appuie sur storage.Uploader pour l'accès Cloudinary

type Service struct {
	uploader storage.Uploader
}

func NewService(u storage.Uploader) *Service {
	return &Service{uploader: u}
}

// UploadProfilePicture uploade une photo de profil selon le rôle
func (s *Service) UploadProfilePicture(ctx context.Context, role models.UserRole, userID string, fh *multipart.FileHeader) (publicID, url string, err error) {
	if err := validateImageHeader(fh); err != nil {
		return "", "", err
	}

	file, tmpPath, err := toTempFile(fh)
	if err != nil {
		return "", "", err
	}
	defer func() { file.Close(); os.Remove(tmpPath) }()

	folder := "livraison/clients/photo_profil"
	if role == models.UserRoleLivreur {
		folder = "livraison/livreurs/photo_profil"
	}
	publicName := fmt.Sprintf("user_%s_%d", userID, time.Now().Unix())
pid, url, err := s.uploader.UploadToFolder(ctx, file, folder, publicName)
	if err != nil {
		return "", "", err
	}
	_, _ = s.persistFileRecord(userID, fh, url, pid, prismadb.FileCategoryProfile)
	return pid, url, nil
}

// UploadClientDocument uploade un document au dossier client
func (s *Service) UploadClientDocument(ctx context.Context, userID string, fh *multipart.FileHeader) (publicID, url string, err error) {
	if err := validateImageHeader(fh); err != nil {
		return "", "", err
	}
	file, tmpPath, err := toTempFile(fh)
	if err != nil {
		return "", "", err
	}
	defer func() { file.Close(); os.Remove(tmpPath) }()

	folder := filepath.ToSlash(fmt.Sprintf("livraison/clients/document/%s", userID))
	publicName := fmt.Sprintf("client_doc_%s_%d", userID, time.Now().Unix())
pid, url, err := s.uploader.UploadToFolder(ctx, file, folder, publicName)
	if err != nil { return "", "", err }
	_, _ = s.persistFileRecord(userID, fh, url, pid, prismadb.FileCategoryClientDoc)
	return pid, url, nil
}

// UploadDriverDocument uploade un document (cni/carte_grise/permis) recto/verso
func (s *Service) UploadDriverDocument(ctx context.Context, userID, docType, side string, fh *multipart.FileHeader) (publicID, url string, err error) {
	docType = strings.ToLower(docType)
	side = strings.ToLower(side)
	allowedTypes := map[string]bool{"cni": true, "carte_grise": true, "permis": true}
	if !allowedTypes[docType] {
		return "", "", fmt.Errorf("Type de document invalide (cni|carte_grise|permis)")
	}
	if side != "recto" && side != "verso" {
		return "", "", fmt.Errorf("Valeur 'side' invalide (recto/verso)")
	}
	if err := validateImageHeader(fh); err != nil {
		return "", "", err
	}

	file, tmpPath, err := toTempFile(fh)
	if err != nil {
		return "", "", err
	}
	defer func() { file.Close(); os.Remove(tmpPath) }()

	folder := filepath.ToSlash(fmt.Sprintf("livraison/livreurs/document/document_vehicule/%s/%s/%s", docType, userID, side))
	publicName := fmt.Sprintf("%s_%s_%s_%d", docType, userID, side, time.Now().Unix())
	pid, url, err := s.uploader.UploadToFolder(ctx, file, folder, publicName)
	if err != nil { return "", "", err }
	// persistance générique
_, _ = s.persistFileRecord(userID, fh, url, pid, func() prismadb.FileCategory {
		if docType == "cni" { return prismadb.FileCategoryDriverCni }
		if docType == "permis" { return prismadb.FileCategoryDriverPermis }
		return prismadb.FileCategoryDriverCarteGrise
	}())
	// mise à jour des champs dédiés CNI/PERMIS dans User
	_ = s.updateUserIdentityDocs(userID, docType, side, url)
	return pid, url, nil
}

// UploadResult représente un résultat d'upload multiple

type UploadResult struct {
	Index    int    `json:"index"`
	PublicID string `json:"publicId"`
	URL      string `json:"url"`
}

// UploadDriverVehicleImages uploade jusqu'à 4 images de véhicule
func (s *Service) UploadDriverVehicleImages(ctx context.Context, userID string, files []*multipart.FileHeader) ([]UploadResult, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("Aucun fichier fourni")
	}
	if len(files) > 4 {
		return nil, fmt.Errorf("Maximum 4 images autorisées")
	}

	folder := filepath.ToSlash(fmt.Sprintf("livraison/livreurs/document/image_vehicule/%s", userID))
	results := make([]UploadResult, 0, len(files))

	for i, fh := range files {
		if err := validateImageHeader(fh); err != nil {
			return nil, fmt.Errorf("Image %d invalide: %w", i+1, err)
		}
		file, tmpPath, err := toTempFile(fh)
		if err != nil {
			return nil, err
		}
publicName := fmt.Sprintf("vehicule_%s_%d_%d", userID, time.Now().Unix(), i+1)
		pid, url, err := s.uploader.UploadToFolder(ctx, file, folder, publicName)
		file.Close(); os.Remove(tmpPath)
		if err != nil {
			return nil, fmt.Errorf("Échec upload image %d: %w", i+1, err)
		}
_, _ = s.persistFileRecord(userID, fh, url, pid, prismadb.FileCategoryVehicleImage)
		results = append(results, UploadResult{Index: i, PublicID: pid, URL: url})
	}
	return results, nil
}

// persistFileRecord enregistre un fichier dans la table File pour l'utilisateur
func (s *Service) persistFileRecord(userID string, fh *multipart.FileHeader, url, publicID string, category prismadb.FileCategory) (string, error) {
	ctx := context.Background()
	mime := sniffContentType(fh)
	size := int(fh.Size)
	orig := fh.Filename
	filename := publicID
rec, err := db.PrismaDB.File.CreateOne(
		prismadb.File.Filename.Set(filename),
		prismadb.File.OriginalName.Set(orig),
		prismadb.File.MimeType.Set(mime),
		prismadb.File.Size.Set(size),
		prismadb.File.Path.Set(url),
		prismadb.File.Category.Set(category),
		prismadb.File.User.Link(prismadb.User.ID.Equals(userID)),
	).Exec(ctx)
	if err != nil { return "", err }
	return rec.ID, nil
}

// validateImageHeader vérifie la taille et le content-type du fichier
func validateImageHeader(fh *multipart.FileHeader) error {
	const maxImageSize = 10 * 1024 * 1024
	if fh.Size <= 0 || fh.Size > maxImageSize {
		return fmt.Errorf("Taille de fichier invalide (max 10MB)")
	}
	f, err := fh.Open()
	if err != nil { return fmt.Errorf("Impossible de lire le fichier: %w", err) }
	defer f.Close()
	var sniff [512]byte
	n, _ := f.Read(sniff[:])
	ctype := http.DetectContentType(sniff[:n])
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
		"image/avif": true,
		"image/gif":  true,
		"image/bmp":  true,
		"image/tiff": true,
		"image/heic": true,
		"image/heif": true,
	}
	if !allowed[strings.ToLower(ctype)] {
		return fmt.Errorf("Type de fichier non pris en charge: %s", ctype)
	}
return nil
}

// updateUserIdentityDocs met à jour les champs dédiés CNI/Permis du User
func (s *Service) updateUserIdentityDocs(userID, docType, side, url string) error {
	if docType != "cni" && docType != "permis" {
		return nil
	}
	ctx := context.Background()
	upd := []prismadb.UserSetParam{}
	if docType == "cni" {
		if side == "recto" {
			upd = append(upd, prismadb.User.CniRecto.Set(url))
		} else if side == "verso" {
			upd = append(upd, prismadb.User.CniVerso.Set(url))
		}
	}
	if docType == "permis" {
		if side == "recto" {
			upd = append(upd, prismadb.User.PermisRecto.Set(url))
		} else if side == "verso" {
			upd = append(upd, prismadb.User.PermisVerso.Set(url))
		}
	}
	if len(upd) == 0 { return nil }
	_, err := db.PrismaDB.User.FindUnique(prismadb.User.ID.Equals(userID)).Update(upd...).Exec(ctx)
	return err
}

// sniffContentType détecte le MIME type d'un FileHeader
func sniffContentType(fh *multipart.FileHeader) string {
	f, err := fh.Open()
	if err != nil { return "application/octet-stream" }
	defer f.Close()
	var sniff [512]byte
	n, _ := f.Read(sniff[:])
	return http.DetectContentType(sniff[:n])
}

// toTempFile copie le multipart dans un fichier temporaire et renvoie un *os.File positionné au début
func toTempFile(fh *multipart.FileHeader) (*os.File, string, error) {
	f, err := fh.Open()
	if err != nil { return nil, "", fmt.Errorf("Impossible de lire le fichier: %w", err) }
	defer f.Close()
	tmp, err := os.CreateTemp("", "upload-*.bin")
	if err != nil { return nil, "", fmt.Errorf("Échec création fichier temporaire: %w", err) }
	if _, err := io.Copy(tmp, f); err != nil {
		tmp.Close(); os.Remove(tmp.Name())
		return nil, "", fmt.Errorf("Échec écriture fichier temporaire: %w", err)
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		logrus.WithError(err).Warn("Impossible de repositionner le fichier temp")
	}
	return tmp, tmp.Name(), nil
}