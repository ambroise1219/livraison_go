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
	UploadProfilePicture(ctx context.Context, file any, filename string) (publicID string, url string, err error)
}

type cloudinaryUploader struct {
	cld    *cloudinary.Cloudinary
	folder string
}

// NewCloudinaryUploader crée un uploader Cloudinary
func NewCloudinaryUploader() (Uploader, error) {
	cfg := config.GetConfig()

	// Pour un preset UNSIGNED, on n'a besoin que du cloud name
	if cfg.CloudinaryCloudName == "" {
		return nil, fmt.Errorf("configuration Cloudinary manquante: CLOUDINARY_CLOUD_NAME requis")
	}

	// Pour un preset UNSIGNED, on utilise quand même l'API key/secret (limitation SDK Go)
	if cfg.CloudinaryAPIKey == "" || cfg.CloudinaryAPISecret == "" {
		return nil, fmt.Errorf("configuration Cloudinary manquante: CLOUDINARY_API_KEY et CLOUDINARY_API_SECRET requis")
	}

	cld, err := cloudinary.NewFromParams(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)
	if err != nil {
		return nil, fmt.Errorf("init cloudinary: %w", err)
	}

	fmt.Printf("✅ CLOUDINARY INIT: cloud_name=%s, preset=photo_profil_livraison\n", cfg.CloudinaryCloudName)
	return &cloudinaryUploader{cld: cld, folder: cfg.CloudinaryFolder}, nil
}

func (u *cloudinaryUploader) UploadProfilePicture(ctx context.Context, file any, filename string) (string, string, error) {
	params := uploader.UploadParams{
		UploadPreset: "photo_profil_livraison",
		// Pas de PublicID pour preset UNSIGNED - Cloudinary le génère automatiquement
		ResourceType: "auto",
	}

	fmt.Printf("🔍 CLOUDINARY DEBUG: Uploading file=%v, filename=%s, preset=photo_profil_livraison\n", file, filename)

	res, err := u.cld.Upload.Upload(ctx, file, params)
	if err != nil {
		fmt.Printf("❌ CLOUDINARY ERROR: %v\n", err)
		return "", "", fmt.Errorf("cloudinary upload error: %w", err)
	}

	fmt.Printf("✅ CLOUDINARY RESPONSE: res=%+v\n", res)
	if res != nil {
		fmt.Printf("📊 CLOUDINARY DETAILS: PublicID=%s, SecureURL=%s, URL=%s\n", res.PublicID, res.SecureURL, res.URL)
	}

	// Vérifier si Cloudinary a renvoyé une erreur applicative
	if res == nil || (res.SecureURL == "" && res.URL == "") {
		return res.PublicID, res.SecureURL, fmt.Errorf("cloudinary empty response: publicId=%s url=%s", res.PublicID, res.SecureURL)
	}
	url := res.SecureURL
	if url == "" {
		url = res.URL
	}
	return res.PublicID, url, nil
}
