package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/db"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

func trySendOTP(base string, phone string) error {
	body := map[string]string{"phone": phone}
	b, _ := json.Marshal(body)
	resp, err := http.Post(base+"/api/v1/auth/otp/send", "application/json", bytes.NewReader(b))
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { dat, _ := io.ReadAll(resp.Body); return fmt.Errorf("send otp %d: %s", resp.StatusCode, string(dat)) }
	return nil
}

func fetchOTPFromDB(phone string) (string, error) {
	ctx := context.Background()
	otp, err := db.PrismaDB.Otp.FindFirst(
		prismadb.Otp.Phone.Equals(phone),
	).OrderBy(
		prismadb.Otp.CreatedAt.Order(prismadb.SortOrderDesc),
	).Exec(ctx)
	if err != nil { return "", err }
	return otp.Code, nil
}

func verifyOTP(base, phone, code string) (string, error) {
	body := map[string]string{"phone": phone, "code": code}
	b, _ := json.Marshal(body)
	resp, err := http.Post(base+"/api/v1/auth/otp/verify", "application/json", bytes.NewReader(b))
	if err != nil { return "", err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { dat, _ := io.ReadAll(resp.Body); return "", fmt.Errorf("verify %d: %s", resp.StatusCode, string(dat)) }
	var out struct { AccessToken string `json:"accessToken"` }
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return "", err }
	if out.AccessToken == "" { return "", fmt.Errorf("no accessToken in response") }
	return out.AccessToken, nil
}

func genPNG() (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 0x22, G: 0x88, B: 0xDD, A: 0xFF})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil { return "", err }
	f, err := os.CreateTemp("", "upload-e2e-*.png")
	if err != nil { return "", err }
	defer f.Close()
	if _, err := f.Write(buf.Bytes()); err != nil { return "", err }
	return f.Name(), nil
}

func uploadProfile(base, token, filePath string) (string, error) {
	bf := &bytes.Buffer{}
	mw := multipart.NewWriter(bf)
	fw, err := mw.CreateFormFile("file", filepath.Base(filePath))
	if err != nil { return "", err }
	f, err := os.Open(filePath)
	if err != nil { return "", err }
	defer f.Close()
	if _, err := io.Copy(fw, f); err != nil { return "", err }
	mw.Close()

	req, _ := http.NewRequest(http.MethodPost, base+"/api/v1/auth/profile/picture", bf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 { return "", fmt.Errorf("upload %d: %s", resp.StatusCode, string(b)) }
	return string(b), nil
}

func main() {
	phone := "+2250700000001"

	// init prisma (to read OTP back)
	_ = config.LoadConfig()
	if err := database.InitPrisma(); err != nil { fmt.Println("ERR:", err); os.Exit(1) }
	defer database.ClosePrisma()
	if err := db.InitializePrisma(); err != nil { fmt.Println("ERR:", err); os.Exit(1) }
	defer db.ClosePrisma()

	bases := []string{"http://localhost:3000", "http://localhost:8080"}
	var base string
	for _, b := range bases {
		if err := trySendOTP(b, phone); err == nil {
			base = b
			break
		}
	}
	if base == "" { fmt.Println("ERR: backend introuvable sur 3000/8080"); os.Exit(1) }
	fmt.Println("BASE:", base)

	// OTP propagation to DB
	time.Sleep(500 * time.Millisecond)
	code, err := fetchOTPFromDB(phone)
	if err != nil { fmt.Println("ERR:", err); os.Exit(1) }
	fmt.Println("OTP:", code)

	token, err := verifyOTP(base, phone, code)
	if err != nil { fmt.Println("ERR:", err); os.Exit(1) }
	fmt.Println("TOKEN:", token[:16]+"...")

	img, err := genPNG()
	if err != nil { fmt.Println("ERR:", err); os.Exit(1) }
	defer os.Remove(img)
	fmt.Println("IMG:", img)

	resp, err := uploadProfile(base, token, img)
	if err != nil { fmt.Println("ERR:", err); os.Exit(1) }
	fmt.Println("UPLOAD RESPONSE:", resp)
}
