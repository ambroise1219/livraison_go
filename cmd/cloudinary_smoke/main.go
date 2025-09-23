package main

import (
    "bytes"
    "context"
    "encoding/base64"
    "fmt"
    "image"
    "image/color"
    "image/png"
    "os"
    "path/filepath"
    "time"

    "github.com/ambroise1219/livraison_go/config"
    "github.com/ambroise1219/livraison_go/services/storage"
)

// generateTinyPNG returns a temp file path containing a 1x1 PNG (solid color)
func generateTinyPNG() (string, error) {
    img := image.NewRGBA(image.Rect(0, 0, 1, 1))
    img.Set(0, 0, color.RGBA{R: 0x22, G: 0x88, B: 0xDD, A: 0xFF})

    var buf bytes.Buffer
    if err := png.Encode(&buf, img); err != nil {
        return "", err
    }

    f, err := os.CreateTemp("", "smoke-*.png")
    if err != nil {
        return "", err
    }
    defer f.Close()

    if _, err := f.Write(buf.Bytes()); err != nil {
        return "", err
    }
    return f.Name(), nil
}

func upload(u storage.Uploader, filePath, folder, publicID string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    f, err := os.Open(filePath)
    if err != nil { return err }
    defer f.Close()

    pid, url, err := u.UploadToFolder(ctx, f, folder, publicID)
    if err != nil {
        return fmt.Errorf("upload failed to %s: %w", folder, err)
    }
    fmt.Printf("OK -> folder=%s publicId=%s url=%s\n", folder, pid, url)
    return nil
}

func main() {
    // Charger la config (.env pris en compte par godotenv)
    cfg := config.LoadConfig()
    // Validation minimale (sans afficher les valeurs)
    if cfg.CloudinaryCloudName == "" || cfg.CloudinaryAPIKey == "" || cfg.CloudinaryAPISecret == "" {
        fmt.Println("ERREUR: variables Cloudinary manquantes (voir CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, CLOUDINARY_API_SECRET)")
        os.Exit(1)
    }

    u, err := storage.NewCloudinaryUploader()
    if err != nil {
        fmt.Printf("ERREUR init Cloudinary: %v\n", err)
        os.Exit(1)
    }

    // Create a tiny image file once
    tmp, err := generateTinyPNG()
    if err != nil {
        fmt.Printf("ERREUR génération PNG: %v\n", err)
        os.Exit(1)
    }
    defer os.Remove(tmp)

    ts := time.Now().Unix()
    user := "smoke_user"

    // 1) Clients/photo_profil
    if err := upload(u, tmp, filepath.ToSlash("livraison/clients/photo_profil"), fmt.Sprintf("client_pp_%s_%d", user, ts)); err != nil { fmt.Println(err); os.Exit(1) }

    // 2) Clients/document
    if err := upload(u, tmp, filepath.ToSlash(fmt.Sprintf("livraison/clients/document/%s", user)), fmt.Sprintf("client_doc_%s_%d", user, ts)); err != nil { fmt.Println(err); os.Exit(1) }

    // 3) Livreurs/photo_profil
    if err := upload(u, tmp, filepath.ToSlash("livraison/livreurs/photo_profil"), fmt.Sprintf("driver_pp_%s_%d", user, ts)); err != nil { fmt.Println(err); os.Exit(1) }

    // 4) Livreur document_vehicule (CNI recto/verso)
    if err := upload(u, tmp, filepath.ToSlash(fmt.Sprintf("livraison/livreurs/document/document_vehicule/cni/%s/recto", user)), fmt.Sprintf("cni_%s_recto_%d", user, ts)); err != nil { fmt.Println(err); os.Exit(1) }
    if err := upload(u, tmp, filepath.ToSlash(fmt.Sprintf("livraison/livreurs/document/document_vehicule/cni/%s/verso", user)), fmt.Sprintf("cni_%s_verso_%d", user, ts)); err != nil { fmt.Println(err); os.Exit(1) }

    // 5) Livreur document_vehicule (carte_grise recto/verso)
    if err := upload(u, tmp, filepath.ToSlash(fmt.Sprintf("livraison/livreurs/document/document_vehicule/carte_grise/%s/recto", user)), fmt.Sprintf("cg_%s_recto_%d", user, ts)); err != nil { fmt.Println(err); os.Exit(1) }
    if err := upload(u, tmp, filepath.ToSlash(fmt.Sprintf("livraison/livreurs/document/document_vehicule/carte_grise/%s/verso", user)), fmt.Sprintf("cg_%s_verso_%d", user, ts)); err != nil { fmt.Println(err); os.Exit(1) }

    // 6) Livreur image_vehicule (une image de test; répéter si besoin jusqu'à 4)
    if err := upload(u, tmp, filepath.ToSlash(fmt.Sprintf("livraison/livreurs/document/image_vehicule/%s", user)), fmt.Sprintf("veh_%s_1_%d", user, ts)); err != nil { fmt.Println(err); os.Exit(1) }

    // Petit succès
    // On encode une courte signature pour tracer ce run (pas de secret)
    sig := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("ok-%d", ts)))
    fmt.Printf("SMOKE UPLOADS TERMINEES: %s\n", sig)
}
