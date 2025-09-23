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

type verifyResp struct {
	AccessToken string `json:"accessToken"`
	User struct {
		ID string `json:"id"`
		Role string `json:"role"`
	} `json:"user"`
}

type uploadResp struct {
	Message string `json:"message"`
	PublicID string `json:"publicId"`
	URL string `json:"url"`
}

func baseURL() (string, error) {
	bases := []string{"http://localhost:3000", "http://localhost:8080"}
	for _, b := range bases {
		resp, err := http.Get(b+"/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return b, nil
		}
	}
	return "", fmt.Errorf("backend introuvable sur 3000/8080")
}

func sendOTP(base, phone string) error {
	b, _ := json.Marshal(map[string]string{"phone": phone})
	resp, err := http.Post(base+"/api/v1/auth/otp/send", "application/json", bytes.NewReader(b))
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { dat, _ := io.ReadAll(resp.Body); return fmt.Errorf("send otp %d: %s", resp.StatusCode, string(dat)) }
	return nil
}

func latestOTP(phone string) (string, error) {
	ctx := context.Background()
	otp, err := db.PrismaDB.Otp.FindFirst(
		prismadb.Otp.Phone.Equals(phone),
	).OrderBy(prismadb.Otp.CreatedAt.Order(prismadb.SortOrderDesc)).Exec(ctx)
	if err != nil { return "", err }
	return otp.Code, nil
}

func verify(base, phone, code string) (*verifyResp, error) {
	b, _ := json.Marshal(map[string]string{"phone": phone, "code": code})
	resp, err := http.Post(base+"/api/v1/auth/otp/verify", "application/json", bytes.NewReader(b))
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { dat, _ := io.ReadAll(resp.Body); return nil, fmt.Errorf("verify %d: %s", resp.StatusCode, string(dat)) }
	var out verifyResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
	return &out, nil
}

func genPNG(name string) (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 0x88, G: 0x77, B: 0xEE, A: 0xFF})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil { return "", err }
	f, err := os.CreateTemp("", name+"-*.png")
	if err != nil { return "", err }
	defer f.Close()
	if _, err := f.Write(buf.Bytes()); err != nil { return "", err }
	return f.Name(), nil
}

func uploadFile(base, token, path, endpoint string) (*uploadResp, error) {
	bf := &bytes.Buffer{}
	mw := multipart.NewWriter(bf)
	fw, err := mw.CreateFormFile("file", filepath.Base(path))
	if err != nil { return nil, err }
	f, err := os.Open(path)
	if err != nil { return nil, err }
	if _, err := io.Copy(fw, f); err != nil { f.Close(); return nil, err }
	f.Close(); mw.Close()

	req, _ := http.NewRequest(http.MethodPost, base+endpoint, bf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	dat, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 { return nil, fmt.Errorf("upload %s %d: %s", endpoint, resp.StatusCode, string(dat)) }
	var out uploadResp
	_ = json.Unmarshal(dat, &out)
	return &out, nil
}

func uploadDriverDoc(base, token, path, docType, side string) (*uploadResp, error) {
	bf := &bytes.Buffer{}
	mw := multipart.NewWriter(bf)
	_ = mw.WriteField("type", docType)
	_ = mw.WriteField("side", side)
	fw, err := mw.CreateFormFile("file", filepath.Base(path))
	if err != nil { return nil, err }
	f, err := os.Open(path)
	if err != nil { return nil, err }
	if _, err := io.Copy(fw, f); err != nil { f.Close(); return nil, err }
	f.Close(); mw.Close()

	req, _ := http.NewRequest(http.MethodPost, base+"/api/v1/auth/driver/document", bf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	dat, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 { return nil, fmt.Errorf("driver doc %d: %s", resp.StatusCode, string(dat)) }
	var out uploadResp
	_ = json.Unmarshal(dat, &out)
	return &out, nil
}

func uploadVehicleImages(base, token, vehicleId string, paths []string) error {
	bf := &bytes.Buffer{}
	mw := multipart.NewWriter(bf)
if vehicleId != "" { _ = mw.WriteField("vehicleId", vehicleId) }
	for _, p := range paths {
		fw, err := mw.CreateFormFile("files", filepath.Base(p))
		if err != nil { return err }
		f, err := os.Open(p)
		if err != nil { return err }
		if _, err := io.Copy(fw, f); err != nil { f.Close(); return err }
		f.Close()
	}
	mw.Close()
	req, _ := http.NewRequest(http.MethodPost, base+"/api/v1/auth/driver/vehicle/images", bf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { dat, _ := io.ReadAll(resp.Body); return fmt.Errorf("vehicle images %d: %s", resp.StatusCode, string(dat)) }
	return nil
}

type createVehicleResp struct { vehicle struct { id string `json:"id"` } }

func createVehicle(base, token, userID string) (string, error) {
	b := []byte(`{"type":"CAR","nom":"TestCar","plaqueImmatriculation":"TEST-123","couleur":"blue","marque":"Brand","modele":"Model"}`)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/users/%s/vehicles", base, userID), bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 { return "", fmt.Errorf("create vehicle %d: %s", resp.StatusCode, string(data)) }
	var out map[string]interface{}
	_ = json.Unmarshal(data, &out)
	if v, ok := out["vehicle"].(map[string]interface{}); ok {
		if id, ok := v["id"].(string); ok { return id, nil }
	}
	return "", fmt.Errorf("vehicle id not found in response: %s", string(data))
}

func main() {
	phone := "+2250700000001"
	_ = config.LoadConfig()
	if err := database.InitPrisma(); err != nil { fmt.Println("ERR:", err); os.Exit(1) }
	defer database.ClosePrisma()
	if err := db.InitializePrisma(); err != nil { fmt.Println("ERR:", err); os.Exit(1) }
	defer db.ClosePrisma()

	base, err := baseURL()
	if err != nil { fmt.Println("ERR:", err); os.Exit(1) }
	fmt.Println("BASE:", base)

	// 1) CLIENT TOKEN
	if err := sendOTP(base, phone); err != nil { fmt.Println("ERR sendOTP:", err); os.Exit(1) }
	time.Sleep(500 * time.Millisecond)
	code, err := latestOTP(phone)
	if err != nil { fmt.Println("ERR otp fetch:", err); os.Exit(1) }
	vr, err := verify(base, phone, code)
	if err != nil { fmt.Println("ERR verify:", err); os.Exit(1) }
	fmt.Println("CLIENT USER:", vr.User.ID, vr.User.Role)

	// 2) TEST client document
	doc, _ := genPNG("client-doc")
	defer os.Remove(doc)
	resp1, err := uploadFile(base, vr.AccessToken, doc, "/api/v1/auth/profile/document")
	if err != nil { fmt.Println("ERR client doc:", err); os.Exit(1) }
	fmt.Println("CLIENT DOC:", resp1.PublicID)

	// 3) PASSER EN LIVREUR
	ctx := context.Background()
	_, err = db.PrismaDB.User.FindUnique(prismadb.User.ID.Equals(vr.User.ID)).Update(prismadb.User.Role.Set(prismadb.UserRoleLivreur)).Exec(ctx)
	if err != nil { fmt.Println("ERR set driver:", err); os.Exit(1) }
	// nouveau OTP pour nouveau token avec role LIVREUR
	if err := sendOTP(base, phone); err != nil { fmt.Println("ERR sendOTP2:", err); os.Exit(1) }
	time.Sleep(500 * time.Millisecond)
	code2, err := latestOTP(phone)
	if err != nil { fmt.Println("ERR otp2:", err); os.Exit(1) }
	vr2, err := verify(base, phone, code2)
	if err != nil { fmt.Println("ERR verify2:", err); os.Exit(1) }
	fmt.Println("DRIVER USER:", vr2.User.ID, vr2.User.Role)

	// 4) DRIVER DOCUMENTS cni recto/verso + carte_grise recto/verso
	cniR, _ := genPNG("cni-recto"); defer os.Remove(cniR)
	cniV, _ := genPNG("cni-verso"); defer os.Remove(cniV)
	cgR, _ := genPNG("cg-recto"); defer os.Remove(cgR)
	cgV, _ := genPNG("cg-verso"); defer os.Remove(cgV)
	if _, err := uploadDriverDoc(base, vr2.AccessToken, cniR, "cni", "recto"); err != nil { fmt.Println("ERR cni recto:", err); os.Exit(1) }
	if _, err := uploadDriverDoc(base, vr2.AccessToken, cniV, "cni", "verso"); err != nil { fmt.Println("ERR cni verso:", err); os.Exit(1) }
	if _, err := uploadDriverDoc(base, vr2.AccessToken, cgR, "carte_grise", "recto"); err != nil { fmt.Println("ERR cg recto:", err); os.Exit(1) }
	if _, err := uploadDriverDoc(base, vr2.AccessToken, cgV, "carte_grise", "verso"); err != nil { fmt.Println("ERR cg verso:", err); os.Exit(1) }
	fmt.Println("DRIVER DOCS: OK")

// 5) CREATE VEHICLE, then VEHICLE IMAGES (2)
	vehId, err := createVehicle(base, vr2.AccessToken, vr2.User.ID)
	if err != nil { fmt.Println("ERR create vehicle:", err); os.Exit(1) }
	fmt.Println("VEHICLE:", vehId)
	v1, _ := genPNG("veh1"); defer os.Remove(v1)
	v2, _ := genPNG("veh2"); defer os.Remove(v2)
	if err := uploadVehicleImages(base, vr2.AccessToken, vehId, []string{v1, v2}); err != nil { fmt.Println("ERR vehicle images:", err); os.Exit(1) }
	fmt.Println("DRIVER VEHICLE IMAGES: OK")
}
