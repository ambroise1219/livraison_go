package storage

import (
	"context"
	"fmt"

	"github.com/ambroise1219/livraison_go/config"
	cloudinary "github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Uploader définit l'interface d'upload de fichiers
type Uploader interface {
	// Legacy: upload photo de profil (conserve pour compat)
	UploadProfilePicture(ctx context.Context, file any, filename string) (publicID string, url string, err error)
	// Générique: uploader vers un sous-dossier précis (recommandé)
	UploadToFolder(ctx context.Context, file any, folder string, publicID string) (publicIDOut string, url string, err error)
}

type cloudinaryUploader struct {
	cld        *cloudinary.Cloudinary
	baseFolder string
	preset     string
}

// NewCloudinaryUploader crée un uploader Cloudinary
func NewCloudinaryUploader() (Uploader, error) {
	cfg := config.GetConfig()

	if cfg.CloudinaryCloudName == "" {
		return nil, fmt.Errorf("configuration Cloudinary manquante: CLOUDINARY_CLOUD_NAME requis")
	}
	if cfg.CloudinaryAPIKey == "" || cfg.CloudinaryAPISecret == "" {
		return nil, fmt.Errorf("configuration Cloudinary manquante: CLOUDINARY_API_KEY et CLOUDINARY_API_SECRET requis")
	}

	cld, err := cloudinary.NewFromParams(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)
	if err != nil {
		return nil, fmt.Errorf("init cloudinary: %w", err)
	}

	preset := "photo_profil_livraison" // preset demandé
	base := cfg.CloudinaryFolder
	if base == "" {
		base = "photo_profil_livraison"
	}
	fmt.Printf("✅ CLOUDINARY INIT: cloud_name=%s, preset=%s, base_folder=%s\n", cfg.CloudinaryCloudName, preset, base)
	return &cloudinaryUploader{cld: cld, baseFolder: base, preset: preset}, nil
}

// UploadProfilePicture conserve une compatabilité, en déléguant vers UploadToFolder
func (u *cloudinaryUploader) UploadProfilePicture(ctx context.Context, file any, filename string) (string, string, error) {
	// Dossier par défaut pour compatibilité (photo_profil sous baseFolder)
	folder := fmt.Sprintf("%s/%s", u.baseFolder, "photo_profil")
	return u.UploadToFolder(ctx, file, folder, filename)
}

// UploadToFolder envoie un fichier dans un sous-dossier donné, avec le preset configuré
func (u *cloudinaryUploader) UploadToFolder(ctx context.Context, file any, folder string, publicID string) (string, string, error) {
	// Pour un preset unsigned imposant un dossier (asset folder), on encode le sous-dossier dans le public_id
	fullPublicID := publicID
	if folder != "" {
		fullPublicID = fmt.Sprintf("%s/%s", folder, publicID)
	}
	params := uploader.UploadParams{
		UploadPreset:  u.preset,
		PublicID:     fullPublicID,
		ResourceType: "image",
	}

	fmt.Printf("🔍 CLOUDINARY DEBUG: Uploading file=%v, folder=%s, publicID=%s, preset=%s\n", file, folder, publicID, u.preset)

	res, err := u.cld.Upload.Upload(ctx, file, params)
	if err != nil {
		fmt.Printf("❌ CLOUDINARY ERROR: %v\n", err)
		return "", "", fmt.Errorf("cloudinary upload error: %w", err)
	}

	if res == nil || (res.SecureURL == "" && res.URL == "") {
		return "", "", fmt.Errorf("cloudinary empty response")
	}
	url := res.SecureURL
	if url == "" {
		url = res.URL
	}
	return res.PublicID, url, nil
}
