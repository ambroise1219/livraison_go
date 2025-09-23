package tests

import (
    "bytes"
    "context"
    "encoding/json"
    "mime/multipart"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/ambroise1219/livraison_go/database"
    "github.com/ambroise1219/livraison_go/db"
    "github.com/ambroise1219/livraison_go/handlers"
    "github.com/ambroise1219/livraison_go/routes"
    prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

func TestDocumentsAndVehicleImages(t *testing.T) {
    // init prisma + handlers
    if err := database.InitPrisma(); err != nil { t.Fatal(err) }
    defer database.ClosePrisma()
    if err := db.InitializePrisma(); err != nil { t.Fatal(err) }
    defer db.ClosePrisma()
    handlers.InitHandlers()

    srv := httptest.NewServer(routes.SetupRoutes())
    defer srv.Close()

    phone := "+2250700000999"
    // send otp
    body := map[string]string{"phone": phone}
    b, _ := json.Marshal(body)
    resp, err := http.Post(srv.URL+"/api/v1/auth/otp/send", "application/json", bytes.NewReader(b))
    if err != nil { t.Fatal(err) }
    resp.Body.Close()

    // get otp from db
    ctx := context.Background()
    otpModel, err := db.PrismaDB.Otp.FindFirst(
        prismadb.Otp.Phone.Equals(phone),
    ).OrderBy(
        prismadb.Otp.CreatedAt.Order(prismadb.SortOrderDesc),
    ).Exec(ctx)
    if err != nil { t.Fatal(err) }
    otp := struct{ Code string }{ Code: otpModel.Code }
    if err != nil { t.Fatal(err) }

    // verify
    vreq := map[string]string{"phone": phone, "code": otp.Code}
    vb, _ := json.Marshal(vreq)
    vresp, err := http.Post(srv.URL+"/api/v1/auth/otp/verify", "application/json", bytes.NewReader(vb))
    if err != nil { t.Fatal(err) }
    var vout struct{ AccessToken string `json:"accessToken"`; User struct{ ID string `json:"id"` } `json:"user"` }
    json.NewDecoder(vresp.Body).Decode(&vout)
    vresp.Body.Close()
    if vout.AccessToken == "" { t.Fatal("no token") }

    // upload client document (multipart) with a valid tiny PNG
    var pngbuf bytes.Buffer
    // 1x1 true PNG header:
    pngbytes := []byte{137,80,78,71,13,10,26,10}
    pngbuf.Write(pngbytes)
    var buf bytes.Buffer
    mw := multipart.NewWriter(&buf)
    fw, _ := mw.CreateFormFile("file", "doc.png")
    fw.Write(pngbuf.Bytes())
    mw.Close()
    req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/auth/profile/document", &buf)
    req.Header.Set("Authorization", "Bearer "+vout.AccessToken)
    req.Header.Set("Content-Type", mw.FormDataContentType())
    r2, err := http.DefaultClient.Do(req)
    if err != nil { t.Fatal(err) }
    r2.Body.Close()

    // list user documents (client)
    reqList, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/auth/documents?type=client", nil)
    reqList.Header.Set("Authorization", "Bearer "+vout.AccessToken)
    r3, err := http.DefaultClient.Do(reqList)
    if err != nil { t.Fatal(err) }
    r3.Body.Close()

    // give time
    time.Sleep(200 * time.Millisecond)
}
