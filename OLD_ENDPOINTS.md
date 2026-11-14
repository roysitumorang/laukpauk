# Old endpoints used by laukpauk_mobile

Every backend call the Android app ([roysitumorang/laukpauk_mobile](https://github.com/roysitumorang/laukpauk_mobile), branch `dev`) makes, taken from its Retrofit interfaces in `android/app/src/main/java/id/laukpauk/data/remote/`. The list reflects the app after its seller and market code was removed (all products belong to a single seller) and after every call moved to JSON bodies, JWT auth and the `X-Version-Code` / `X-Request-Id` headers ([laukpauk_mobile#12](https://github.com/roysitumorang/laukpauk_mobile/pull/12)). Sign-in is `POST users/sign_in` ([laukpauk_mobile#13](https://github.com/roysitumorang/laukpauk_mobile/pull/13)). These are the endpoints the app calls, not the endpoints this repository serves; use this list to map the app onto the current API.

## Conventions

- **Base URL:** paths are relative to `API_BASE_URL` (a build setting, e.g. `https://<host>/api/`).
- **Headers on every backend request** (the examples below leave the first two out):
  - `X-Version-Code`: the app's version code (build number).
  - `X-Request-Id`: a random UUID v4, new for each request.
  - `Accept: application/json`.
- **Auth:** JWT. While a user is signed in, every backend request carries `Authorization: Bearer <JWT>` (the `id_token` returned by `POST users/sign_in`, until its `expired_at` passes). The examples show the header on the calls that need a signed-in user; no call sends the token in the URL or the body.
- **Bodies:** every `POST` sends `Content-Type: application/json` with a JSON object, except `Transaksi/pesananDiterima/{id}`, which is sent with no body. Keys whose value is null are left out. Most values are JSON strings, because the app holds them as text. The profile photo is sent as a base64 string.
- **Path parameters** are shown as `{name}`; placeholder values as `<name>`.
- **Responses** are JSON. The examples list only the fields the app reads (from its response classes), with type placeholders: `"string"`, `0`, `0.0`, `true`.
- **Overloads:** where the app calls a path both with and without a trailing `/{id}`, both forms are listed.

## Contents

- [AccountApi](#accountapi) (5)
- [AuthApi](#authapi) (7)
- [CartApi](#cartapi) (10)
- [CheckoutApi](#checkoutapi) (14)
- [DeliveryApi](#deliveryapi) (5)
- [HomeApi](#homeapi) (11)
- [NotifikasiApi](#notifikasiapi) (7)
- [PaketApi](#paketapi) (3)
- [PaymentApi](#paymentapi) (2)
- [ProdukApi](#produkapi) (6)
- [PromoApi](#promoapi) (2)
- [PushApi](#pushapi) (2)
- [RepeatOrderApi](#repeatorderapi) (2)
- [ResepApi](#resepapi) (5)
- [SearchApi](#searchapi) (1)
- [SettingApi](#settingapi) (2)
- [TransaksiApi](#transaksiapi) (8)
- [UserApi](#userapi) (9)
- [VoucherApi](#voucherapi) (2)

103 endpoint definitions in total.

## Index

| Method | Path | JWT required | App method |
|---|---|---|---|
| GET | `Auth/removeToken` | yes | `AccountApi.removeToken` |
| POST | `Auth/gantiNomor` | yes | `AccountApi.gantiNomor` |
| POST | `Auth/verifikasiNomor` | yes | `AccountApi.verifikasiNomor` |
| GET | `Auth/jumlahGantiNomorHPForToday` | yes | `AccountApi.jumlahGantiNomorHPForToday` |
| GET | `Pengaturan/getReferralConfiguration` | yes | `AccountApi.getReferralConfiguration` |
| POST | `users/sign_in` | no | `AuthApi.signIn` |
| POST | `Auth/register` | no | `AuthApi.register` |
| POST | `Auth/konfirmasiOTP` | no | `AuthApi.konfirmasiOTP` |
| POST | `Auth/verifikasiUser` | no | `AuthApi.verifikasiUser` |
| POST | `Auth/resendOTP` | yes | `AuthApi.resendOTP` |
| POST | `Auth/lupaPassword` | no | `AuthApi.lupaPassword` |
| GET | `User/getDetailUser` | yes | `AuthApi.getDetailUser` |
| GET | `User/getJumlahCart` | no | `CartApi.getJumlahCart` |
| GET | `User/getCart` | no | `CartApi.getCart` |
| GET | `User/getTotalKeranjang` | no | `CartApi.getTotalKeranjang` |
| GET | `Keranjang/bersihkanKeranjang` | no | `CartApi.bersihkanKeranjang` |
| POST | `Keranjang/saveCart` | no | `CartApi.saveCart` |
| POST | `Keranjang/saveCart/{idItem}` | no | `CartApi.updateCart` |
| POST | `Keranjang/deleteCart/{idCart}` | no | `CartApi.deleteCart` |
| GET | `Pengaturan/getWalletConfiguration` | no | `CartApi.getWalletConfiguration` |
| GET | `gift/gift-transaksi` | no | `CartApi.getGiftTransaksi` |
| GET | `Gift/getGift/{idGift}` | no | `CartApi.getGift` |
| GET | `Auth/cekVerifikasiUser` | yes | `CheckoutApi.cekVerifikasiUser` |
| GET | `User/getDetailUser` | yes | `CheckoutApi.getDetailUser` |
| POST | `JadwalDelivery/isJadwalDeliveryOpened` | no | `CheckoutApi.isJadwalDeliveryOpened` |
| GET | `User/getTotalKeranjang` | no | `CheckoutApi.getTotalKeranjang` |
| GET | `User/getTotalOngkir` | no | `CheckoutApi.getTotalOngkir` |
| GET | `User/getJumlahCart` | no | `CheckoutApi.getJumlahCart` |
| GET | `Transaksi/getJenisOngkir` | no | `CheckoutApi.getJenisOngkir` |
| GET | `User/getDiskonKategori` | no | `CheckoutApi.getDiskonKategori` |
| GET | `Pengaturan/getWalletConfiguration` | no | `CheckoutApi.getWalletConfiguration` |
| GET | `User/getTotalDompet` | yes | `CheckoutApi.getTotalDompet` |
| GET | `Point/getPoint` | yes | `CheckoutApi.getPoint` |
| GET | `User/getVoucher` | yes | `CheckoutApi.getVoucher` |
| GET | `Voucher/getStatusVoucher/{id}` | yes | `CheckoutApi.getStatusVoucher` |
| GET | `Promo/cekKodePromo/{kode}` | yes | `CheckoutApi.cekKodePromo` |
| GET | `JadwalDelivery/getAvailableJadwalDelivery` | no | `DeliveryApi.getAvailableJadwalDelivery` |
| GET | `Gift/getGift` | no | `DeliveryApi.getGifts` |
| GET | `Keranjang/bersihkanKeranjang` | no | `DeliveryApi.bersihkanKeranjang` |
| GET | `https://maps.googleapis.com/maps/api/geocode/json` | no | `DeliveryApi.reverseGeocode` |
| GET | `https://maps.googleapis.com/maps/api/place/textsearch/json` | no | `DeliveryApi.searchPlaces` |
| GET | `Banner/getBanner` | no | `HomeApi.getBanner` |
| GET | `Artikel/getArtikel` | no | `HomeApi.getArtikel` |
| GET | `User/getJumlahVoucher` | yes | `HomeApi.getJumlahVoucher` |
| GET | `User/getTotalDompet` | yes | `HomeApi.getTotalDompet` |
| GET | `Point/getPoint` | yes | `HomeApi.getPoint` |
| GET | `Produk/getProduk` | yes | `HomeApi.getProduk` |
| GET | `Konten/getKonten` | no | `HomeApi.getKonten` |
| GET | `Konten/getKontenItemFromThisKonten/{idKonten}` | no | `HomeApi.getKontenItem` |
| GET | `KategoriProduk/getKategoriProduk` | no | `HomeApi.getKategoriProduk` |
| GET | `Produk/getProductsInThisCategory/{idKategoriProduk}` | no | `HomeApi.getProdukInThisCategory` |
| GET | `Produk/getLowestPriceOfThisProduct/{idProduk}` | no | `HomeApi.getLowestPrice` |
| GET | `User/getNotifikasi` | yes | `NotifikasiApi.getNotifikasi` |
| GET | `User/getNotifikasiKhusus` | yes | `NotifikasiApi.getNotifikasiKhusus` |
| GET | `User/getNotifikasiKhusus/{idUser}` | yes | `NotifikasiApi.getNotifikasiKhusus` |
| GET | `User/getJumlahNotifikasi` | yes | `NotifikasiApi.getJumlahNotifikasi` |
| GET | `User/getJumlahNotifikasiKhusus` | yes | `NotifikasiApi.getJumlahNotifikasiKhusus` |
| GET | `Notifikasi/getNotifikasi/{idNotifikasi}` | yes | `NotifikasiApi.getDetailNotifikasi` |
| POST | `Notifikasi/setAs` | yes | `NotifikasiApi.setAs` |
| GET | `Produk/detailPaket` | no | `PaketApi.detailPaket` |
| GET | `Produk/getPaketLainnya` | no | `PaketApi.getPaketLainnya` |
| POST | `Keranjang/saveCart` | no | `PaketApi.saveCart` |
| GET | `Xendit/getPaymentChannels` | yes | `PaymentApi.getPaymentChannels` |
| POST | `Transaksi/checkout` | yes | `PaymentApi.checkout` |
| GET | `Produk/getProduk` | yes | `ProdukApi.getProduk` |
| GET | `Produk/detailProduk` | no | `ProdukApi.detailProduk` |
| GET | `Produk/getLowestPriceOfThisProduct/{idProduk}` | no | `ProdukApi.getLowestPrice` |
| GET | `Produk/getHargaProduk` | no | `ProdukApi.getHargaProduk` |
| GET | `Produk/getProdukLainnya` | no | `ProdukApi.getProdukLainnya` |
| GET | `Produk/isThisProductInCA` | no | `ProdukApi.isThisProductInCA` |
| GET | `Promo/getPromoAroundOfThisLatLong` | yes | `PromoApi.getPromoAroundOfThisLatLong` |
| GET | `Promo/getPromoItem/{idPromoItem}` | no | `PromoApi.getPromoItem` |
| POST | `User/registerFcmToken` | yes | `PushApi.registerFcmToken` |
| POST | `User/unregisterFcmToken` | yes | `PushApi.unregisterFcmToken` |
| GET | `RO/getItemTransaksi` | yes | `RepeatOrderApi.getItemTransaksi` |
| POST | `RO/addToCart` | yes | `RepeatOrderApi.addToCart` |
| GET | `Resep/getResep` | no | `ResepApi.getResep` |
| GET | `Resep/getResep/{idResep}` | no | `ResepApi.getDetailResep` |
| GET | `Resep/getResepItemFromThisResep/{idResep}` | no | `ResepApi.getResepItem` |
| GET | `Produk/getHargaProduk` | no | `ResepApi.getHargaProduk` |
| POST | `Keranjang/belanjaResep` | no | `ResepApi.belanjaResep` |
| GET | `Konten/pencarian` | no | `SearchApi.pencarian` |
| GET | `Pengaturan/getSettingValue/{key}` | no | `SettingApi.getSettingValue` |
| GET | `Pengaturan/getContactUsWhatsappConfiguration` | no | `SettingApi.getContactUsWhatsappConfiguration` |
| GET | `User/getRiwayatTransaksi` | yes | `TransaksiApi.getRiwayatTransaksi` |
| GET | `Transaksi/getTransaksi/{id}` | yes | `TransaksiApi.getTransaksi` |
| GET | `Transaksi/getItemTransaksiFromTransaksi/{id}` | yes | `TransaksiApi.getItemTransaksi` |
| GET | `Transaksi/getJumlahItemDiantar/{id}` | yes | `TransaksiApi.getJumlahItemDiantar` |
| POST | `Transaksi/pesananDiterima/{id}` | yes | `TransaksiApi.pesananDiterima` |
| POST | `Transaksi/belanjaUlang/{id}` | no | `TransaksiApi.belanjaUlang` |
| GET | `gift/gift-transaksi` | no | `TransaksiApi.getGiftTransaksi` |
| GET | `Gift/getGift/{idGift}` | no | `TransaksiApi.getGift` |
| GET | `User/getDetailUser` | yes | `UserApi.getDetailUser` |
| GET | `User/getTotalDompet` | yes | `UserApi.getTotalDompet` |
| GET | `User/getJumlahVoucher` | yes | `UserApi.getJumlahVoucher` |
| GET | `Point/getPoint` | yes | `UserApi.getPoint` |
| GET | `User/getRiwayatDompet` | yes | `UserApi.getRiwayatDompet` |
| GET | `User/getRiwayatPoint` | yes | `UserApi.getRiwayatPoint` |
| POST | `User/updateProfil` | yes | `UserApi.updateProfil` |
| POST | `User/updateFotoProfil` | yes | `UserApi.updateFotoProfil` |
| POST | `User/gantiPassword` | yes | `UserApi.gantiPassword` |
| GET | `User/getVoucher` | yes | `VoucherApi.getVoucher` |
| GET | `Voucher/getVoucher/{idVoucher}` | yes | `VoucherApi.getVoucherDetail` |

## AccountApi

### `GET Auth/removeToken`

App method: `AccountApi.removeToken`.

Request:

```http
GET Auth/removeToken?cartCode=<cartCode>
Authorization: Bearer <JWT>
```

Response (`AccountResponses.RemoveToken`):

```json
{
  "removeToken": true
}
```

### `POST Auth/gantiNomor`

App method: `AccountApi.gantiNomor`.

Request:

```http
POST Auth/gantiNomor
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "nomorBaru": "<nomorBaru>"
}
```

Response (`AccountResponses.GantiNomor`):

```json
{
  "statusSimpanNomorBaru": true,
  "idPenggantianNomor": 0
}
```

### `POST Auth/verifikasiNomor`

App method: `AccountApi.verifikasiNomor`.

Request:

```http
POST Auth/verifikasiNomor
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "idPenggantianNomor": "<idPenggantianNomor>",
  "kodeVerifikasi": "<kodeVerifikasi>"
}
```

Response (`AccountResponses.VerifikasiNomor`):

```json
{
  "statusVerifikasiNomor": true
}
```

### `GET Auth/jumlahGantiNomorHPForToday`

App method: `AccountApi.jumlahGantiNomorHPForToday`.

Request:

```http
GET Auth/jumlahGantiNomorHPForToday
Authorization: Bearer <JWT>
```

Response (`AccountResponses.JumlahGantiNomor`):

```json
{
  "jumlahGantiNomorHPForToday": 0
}
```

### `GET Pengaturan/getReferralConfiguration`

App method: `AccountApi.getReferralConfiguration`.

Request:

```http
GET Pengaturan/getReferralConfiguration
Authorization: Bearer <JWT>
```

Response (`AccountResponses.ReferralConfiguration`):

```json
{
  "referralConfiguration": {
    "referralWallet": 0
  }
}
```

## AuthApi

### `POST users/sign_in`

App method: `AuthApi.signIn`.

Sign-in. Body: `login` (string) and `password` (string, base64 of the UTF-8 password). Answers `id_token` (the JWT), `expired_at` (RFC 3339 date-time) and `user` (the signed-in user), either at the top level or inside the backend's `data` envelope. Wrong credentials answer a 4xx status.

Request:

```http
POST users/sign_in
Content-Type: application/json

{
  "login": "<phone number>",
  "password": "<base64 of the password>"
}
```

Response (`id_token`, `expired_at`, `user`):

```json
{
  "data": {
    "id_token": "<JWT>",
    "expired_at": "2026-11-15T21:09:00+07:00",
    "user": {
      "id": "<uuid>",
      "role": "buyer",
      "name": "string",
      "mobile_phone": "string",
      "mobile_phone_confirmed_at": "2026-11-14T21:09:00+07:00",
      "username": "string",
      "email": "string",
      "photo": "string",
      "user_level": "string",
      "referral_code": "string",
      "address": "string"
    }
  }
}
```

The app reads the three attributes from `data` (the backend's response envelope) when it is present, and from the top level otherwise. It stores `id_token` as the session JWT, `expired_at` (RFC 3339 date-time) as its expiry, and `user` as the current user. From `user` it reads `id` (the user id) and `mobile_phone_confirmed_at` (null opens the phone-verification screen); the other `user` fields shown are the ones the backend's user model sends, and are stored as they come. After `expired_at` the app treats the session as signed out and stops sending the token. A 4xx status means the login or password is wrong. In the request, `password` is always the standard base64 (with padding) of the password's UTF-8 bytes, for example `cmFoYXNpYTEyMw==` for `rahasia123`; `Auth/register` and `User/gantiPassword` still send their passwords as typed.

### `POST Auth/register`

App method: `AuthApi.register`.

Request:

```http
POST Auth/register
Content-Type: application/json

{
  "namaLengkap": "<namaLengkap>",
  "nomorHP": "<nomorHP>",
  "password": "<password>",
  "konfirmasiPassword": "<konfirmasiPassword>",
  "kodeReferal": "<kodeReferal>"
}
```

Response (`AuthResponses.Register`):

```json
{
  "statusRegister": true,
  "returnData": {
    "idUser": "string",
    "message": "string",
    "authToken": "string"
  }
}
```

### `POST Auth/konfirmasiOTP`

App method: `AuthApi.konfirmasiOTP`.

Request:

```http
POST Auth/konfirmasiOTP
Content-Type: application/json

{
  "idUser": "<idUser>",
  "otp": "<otp>"
}
```

Response (`AuthResponses.KonfirmasiOTP`):

```json
{
  "statusKonfirmasiOTP": true,
  "returnData": {
    "idUser": "string",
    "message": "string",
    "authToken": "string"
  }
}
```

### `POST Auth/verifikasiUser`

App method: `AuthApi.verifikasiUser`.

Request:

```http
POST Auth/verifikasiUser
Content-Type: application/json

{
  "idUser": "<idUser>",
  "verificationCode": "<verificationCode>"
}
```

Response (`AuthResponses.VerifikasiUser`):

```json
{
  "statusVerifikasiUser": true
}
```

### `POST Auth/resendOTP`

App method: `AuthApi.resendOTP`.

Request:

```http
POST Auth/resendOTP
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "nomorTelepon": "<nomorTelepon>"
}
```

Response (`AuthResponses.ResendOTP`):

```json
{
  "statusResendOTP": true,
  "messageResendOTP": "string"
}
```

### `POST Auth/lupaPassword`

App method: `AuthApi.lupaPassword`.

Request:

```http
POST Auth/lupaPassword
Content-Type: application/json

{
  "nomorHP": "<nomorHP>"
}
```

Response (`AuthResponses.LupaPassword`):

```json
{
  "statusRecoveryPassword": true,
  "returnData": {
    "idUser": "string",
    "message": "string",
    "authToken": "string"
  }
}
```

### `GET User/getDetailUser`

App method: `AuthApi.getDetailUser`.

Request:

```http
GET User/getDetailUser?getUserEntities=<getUserEntities>
Authorization: Bearer <JWT>
```

Response (`AuthResponses.DetailUserResponse`):

```json
{
  "detailUser": {
    "nomorHP": "string"
  }
}
```

## CartApi

### `GET User/getJumlahCart`

App method: `CartApi.getJumlahCart`.

Request:

```http
GET User/getJumlahCart?cartCode=<cartCode>
```

Response (`CartResponses.Count`):

```json
{
  "jumlahKeranjang": 0
}
```

### `GET User/getCart`

App method: `CartApi.getCart`.

Request:

```http
GET User/getCart?cartCode=<cartCode>
```

Response (`CartResponses.ProductList`):

```json
{
  "listCart": [
    {
      "idCart": "string",
      "idProduk": "string",
      "jenisProduk": "string",
      "idSatuan": "string",
      "namaSatuan": "string",
      "jenisSatuan": "string",
      "namaProduk": "string",
      "fotoProduk": "string",
      "harga": "string",
      "diskon": "string",
      "quantity": "string",
      "catatan": "string",
      "idPromo": "string"
    }
  ]
}
```

### `GET User/getTotalKeranjang`

App method: `CartApi.getTotalKeranjang`.

Request:

```http
GET User/getTotalKeranjang?cartCode=<cartCode>&latDelivery=<latDelivery>&longDelivery=<longDelivery>
```

Response (`CartResponses.Total`):

```json
{
  "totalKeranjang": 0
}
```

### `GET Keranjang/bersihkanKeranjang`

App method: `CartApi.bersihkanKeranjang`.

Request:

```http
GET Keranjang/bersihkanKeranjang?cartCode=<cartCode>
```

Response (`CartResponses.Clear`):

```json
{
  "statusBersihkanKeranjang": true
}
```

### `POST Keranjang/saveCart`

App method: `CartApi.saveCart`.

Request:

```http
POST Keranjang/saveCart
Content-Type: application/json

{
  "cartCode": "<cartCode>",
  "idProduk": "<idProduk>",
  "jenisProduk": "nonpaket",
  "idSatuan": "<idSatuan>",
  "quantity": "1",
  "jadwalDeliveryCode": "<jadwalDeliveryCode>",
  "dayDelivery": "Hari Pengantaran",
  "idPromo": "<idPromo>",
  "idUser": "<idUser>"
}
```

Conditional fields: `idPromo` (optional); `idUser` (when logged in).

Response (`AddToCartResult`):

```json
{
  "statusAddCart": true,
  "message": "string"
}
```

### `POST Keranjang/saveCart/{idItem}`

App method: `CartApi.updateCart`.

Request:

```http
POST Keranjang/saveCart/{idItem}
Content-Type: application/json

{
  "quantity": "1",
  "catatan": "<catatan>"
}
```

Response (`AddToCartResult`):

```json
{
  "statusAddCart": true,
  "message": "string"
}
```

### `POST Keranjang/deleteCart/{idCart}`

App method: `CartApi.deleteCart`.

The body is an empty JSON object.

Request:

```http
POST Keranjang/deleteCart/{idCart}
Content-Type: application/json

{}
```

Response (`CartResponses.Delete`):

```json
{
  "statusDeleteCart": true
}
```

### `GET Pengaturan/getWalletConfiguration`

App method: `CartApi.getWalletConfiguration`.

Duplicate of the checkout slice's wallet configuration call.

Request:

```http
GET Pengaturan/getWalletConfiguration
```

Response (`CartResponses.Wallet`):

```json
{
  "walletConfiguration": {
    "isWalletOn": true,
    "walletDiscount": "string",
    "minGrandTotalTransactionForWallet": "string"
  }
}
```

### `GET gift/gift-transaksi`

App method: `CartApi.getGiftTransaksi`.

Request:

```http
GET gift/gift-transaksi?cartCode=<cartCode>
```

Response (`CartResponses.GiftTransaksiList`):

```json
{
  "giftTransaksi": [
    {
      "id": "string",
      "idGift": "string"
    }
  ]
}
```

### `GET Gift/getGift/{idGift}`

App method: `CartApi.getGift`.

Request:

```http
GET Gift/getGift/{idGift}
```

Response (`CartResponses.GiftDetail`):

```json
{
  "gift": {
    "id": "string",
    "nama": "string",
    "deskripsi": "string",
    "foto": "string"
  }
}
```

## CheckoutApi

### `GET Auth/cekVerifikasiUser`

App method: `CheckoutApi.cekVerifikasiUser`.

Request:

```http
GET Auth/cekVerifikasiUser
Authorization: Bearer <JWT>
```

Response (`CheckoutResponses.Verified`):

```json
{
  "isVerified": true
}
```

### `GET User/getDetailUser`

App method: `CheckoutApi.getDetailUser`.

Request:

```http
GET User/getDetailUser?getUserEntities=<select>
Authorization: Bearer <JWT>
```

Response (`CheckoutResponses.DetailUserResponse`):

```json
{
  "detailUser": {
    "nama": "string",
    "nomorHP": "string",
    "alamat": "string",
    "statusOTP": "string"
  }
}
```

### `POST JadwalDelivery/isJadwalDeliveryOpened`

App method: `CheckoutApi.isJadwalDeliveryOpened`.

Request:

```http
POST JadwalDelivery/isJadwalDeliveryOpened
Content-Type: application/json

{
  "kode": "<kode>",
  "day": "<day>",
  "setOn": "<setOn>"
}
```

Response (`CheckoutResponses.JadwalOpened`):

```json
{
  "isJadwalDeliveryOpened": true
}
```

### `GET User/getTotalKeranjang`

App method: `CheckoutApi.getTotalKeranjang`.

Request:

```http
GET User/getTotalKeranjang?cartCode=<cartCode>&latDelivery=<lat>&longDelivery=<lng>
```

Response (`CheckoutResponses.TotalKeranjang`):

```json
{
  "totalKeranjang": 0
}
```

### `GET User/getTotalOngkir`

App method: `CheckoutApi.getTotalOngkir`.

Request:

```http
GET User/getTotalOngkir?cartCode=<cartCode>&latDelivery=<lat>&longDelivery=<lng>
```

Response (`CheckoutResponses.TotalOngkir`):

```json
{
  "totalOngkir": 0
}
```

### `GET User/getJumlahCart`

App method: `CheckoutApi.getJumlahCart`.

Request:

```http
GET User/getJumlahCart?cartCode=<cartCode>
```

Response (`CheckoutResponses.JumlahCart`):

```json
{
  "jumlahKeranjang": 0
}
```

### `GET Transaksi/getJenisOngkir`

App method: `CheckoutApi.getJenisOngkir`.

Request:

```http
GET Transaksi/getJenisOngkir?cartCode=<cartCode>&latDelivery=<lat>&longDelivery=<lng>
```

Response (`CheckoutResponses.JenisOngkir`):

```json
{
  "jenisOngkir": "string",
  "jarak": 0.0
}
```

### `GET User/getDiskonKategori`

App method: `CheckoutApi.getDiskonKategori`.

Request:

```http
GET User/getDiskonKategori?cartCode=<cartCode>
```

Response (`CheckoutResponses.DiskonKategori`):

```json
{
  "diskonKategori": "<any JSON>"
}
```

### `GET Pengaturan/getWalletConfiguration`

App method: `CheckoutApi.getWalletConfiguration`.

Request:

```http
GET Pengaturan/getWalletConfiguration
```

Response (`CheckoutResponses.WalletConfig`):

```json
{
  "walletConfiguration": {
    "isWalletOn": true,
    "minGrandTotalTransactionForWallet": 0,
    "walletDiscount": 0
  }
}
```

### `GET User/getTotalDompet`

App method: `CheckoutApi.getTotalDompet`.

Request:

```http
GET User/getTotalDompet
Authorization: Bearer <JWT>
```

Response (`CheckoutResponses.TotalDompet`):

```json
{
  "totalDompet": 0
}
```

### `GET Point/getPoint`

App method: `CheckoutApi.getPoint`.

Request:

```http
GET Point/getPoint?point=<point>
Authorization: Bearer <JWT>
```

Response (`CheckoutResponses.PointResponse`):

```json
{
  "response": {
    "point": 0
  }
}
```

### `GET User/getVoucher`

App method: `CheckoutApi.getVoucher`.

Request:

```http
GET User/getVoucher?entitasVoucher=<entitasVoucher>&forThisTotalBelanja=<totalBelanja>
Authorization: Bearer <JWT>
```

Response (`CheckoutResponses.VoucherList`):

```json
{
  "listVoucher": [
    {
      "id": "string",
      "nama": "string",
      "foto": "string",
      "minimumBelanja": 0,
      "nilai": 0,
      "jenis": "string",
      "kode": "string",
      "maximumUsage": "string"
    }
  ]
}
```

### `GET Voucher/getStatusVoucher/{id}`

App method: `CheckoutApi.getStatusVoucher`.

Request:

```http
GET Voucher/getStatusVoucher/{id}
Authorization: Bearer <JWT>
```

Response (`CheckoutResponses.StatusVoucher`):

```json
{
  "statusVoucher": "string"
}
```

### `GET Promo/cekKodePromo/{kode}`

App method: `CheckoutApi.cekKodePromo`.

Request:

```http
GET Promo/cekKodePromo/{kode}
Authorization: Bearer <JWT>
```

Response (`CheckoutResponses.KodePromo`):

```json
{
  "status": true,
  "message": "string",
  "data": {
    "promo": {
      "diskon": 0
    }
  }
}
```

## DeliveryApi

### `GET JadwalDelivery/getAvailableJadwalDelivery`

App method: `DeliveryApi.getAvailableJadwalDelivery`.

Request:

```http
GET JadwalDelivery/getAvailableJadwalDelivery
```

Response (`DeliveryResponses.Jadwal`):

```json
{
  "availableJadwalDelivery": [
    {
      "<key>": "<value>"
    }
  ]
}
```

### `GET Gift/getGift`

App method: `DeliveryApi.getGifts`.

Request:

```http
GET Gift/getGift
```

Response (`DeliveryResponses.Gifts`):

```json
{
  "gift": [
    {
      "id": 0,
      "foto": "string",
      "nama": "string"
    }
  ]
}
```

### `GET Keranjang/bersihkanKeranjang`

App method: `DeliveryApi.bersihkanKeranjang`.

Legacy CheckoutModel.bersihkanKeranjang (the response is ignored).

Request:

```http
GET Keranjang/bersihkanKeranjang?cartCode=<cartCode>
```

Response (`Map<String, Object>`):

```json
{
  "<key>": "<value>"
}
```

### `GET https://maps.googleapis.com/maps/api/geocode/json`

App method: `DeliveryApi.reverseGeocode`.

External URL, not the backend.

Request:

```http
GET https://maps.googleapis.com/maps/api/geocode/json?latlng=<latlng>&key=<key>
```

Response (`DeliveryResponses.Geocode`):

```json
{
  "results": [
    {
      "formatted_address": "string"
    }
  ]
}
```

### `GET https://maps.googleapis.com/maps/api/place/textsearch/json`

App method: `DeliveryApi.searchPlaces`.

External URL, not the backend.

Request:

```http
GET https://maps.googleapis.com/maps/api/place/textsearch/json?query=<query>&key=<key>
```

Response (`DeliveryResponses.PlaceSearch`):

```json
{
  "results": [
    {
      "name": "string",
      "formatted_address": "string",
      "geometry": {
        "location": {
          "lat": 0.0,
          "lng": 0.0
        }
      }
    }
  ]
}
```

## HomeApi

### `GET Banner/getBanner`

App method: `HomeApi.getBanner`.

Request:

```http
GET Banner/getBanner
```

Response (`HomeResponses.Banners`):

```json
{
  "listBanner": [
    {
      "id": "string",
      "gambar": "string",
      "nama": "string",
      "keterangan": "string"
    }
  ]
}
```

### `GET Artikel/getArtikel`

App method: `HomeApi.getArtikel`.

Request:

```http
GET Artikel/getArtikel
```

Response (`HomeResponses.Artikels`):

```json
{
  "listArtikel": [
    {
      "judul": "string",
      "foto": "string",
      "link": "string"
    }
  ]
}
```

### `GET User/getJumlahVoucher`

App method: `HomeApi.getJumlahVoucher`.

Request:

```http
GET User/getJumlahVoucher
Authorization: Bearer <JWT>
```

Response (`HomeResponses.JumlahVoucher`):

```json
{
  "jumlahVoucher": 0
}
```

### `GET User/getTotalDompet`

App method: `HomeApi.getTotalDompet`.

Request:

```http
GET User/getTotalDompet
Authorization: Bearer <JWT>
```

Response (`HomeResponses.TotalDompet`):

```json
{
  "totalDompet": 0
}
```

### `GET Point/getPoint`

App method: `HomeApi.getPoint`.

Request:

```http
GET Point/getPoint?point=<point>
Authorization: Bearer <JWT>
```

Response (`HomeResponses.Point`):

```json
{
  "response": {
    "point": 0
  }
}
```

### `GET Produk/getProduk`

App method: `HomeApi.getProduk`.

The first page of the product list (see ProdukApi.getProduk).

Request:

```http
GET Produk/getProduk?jadwalDeliveryCode=<jadwalDeliveryCode>&startFrom=<startFrom>&useLimit=<useLimit>
Authorization: Bearer <JWT>
```

Response (`HomeResponses.ProdukList`):

```json
{
  "listProduk": [
    {
      "id": "string",
      "nama": "string",
      "foto": "string",
      "namaKategoriProduk": "string",
      "keteranganSingkat": "string",
      "stokAvailable": true,
      "promo": "<any JSON>"
    }
  ]
}
```

### `GET Konten/getKonten`

App method: `HomeApi.getKonten`.

Request:

```http
GET Konten/getKonten?page=<page>
```

Response (`HomeResponses.KontenList`):

```json
{
  "dataKonten": [
    {
      "id": "string",
      "namaDivisi": "string",
      "jenisItem": "string",
      "route": "string"
    }
  ]
}
```

### `GET Konten/getKontenItemFromThisKonten/{idKonten}`

App method: `HomeApi.getKontenItem`.

Request:

```http
GET Konten/getKontenItemFromThisKonten/{idKonten}?limit=<limit>
```

Response (`HomeResponses.KontenItems`):

```json
{
  "kontenItem": [
    {
      "pilihan": "string",
      "item": "string",
      "keterangan": "string"
    }
  ]
}
```

### `GET KategoriProduk/getKategoriProduk`

App method: `HomeApi.getKategoriProduk`.

Request:

```http
GET KategoriProduk/getKategoriProduk
```

Response (`HomeResponses.KategoriList`):

```json
{
  "kategoriProduk": [
    {
      "id": "string",
      "nama": "string",
      "gambar": "string"
    }
  ]
}
```

### `GET Produk/getProductsInThisCategory/{idKategoriProduk}`

App method: `HomeApi.getProdukInThisCategory`.

Request:

```http
GET Produk/getProductsInThisCategory/{idKategoriProduk}
```

Response (`HomeResponses.ProdukInKategori`):

```json
{
  "productInThisCategory": [
    {
      "id": "string",
      "nama": "string",
      "foto": "string",
      "namaKategoriProduk": "string",
      "keteranganSingkat": "string",
      "stokAvailable": true,
      "promo": "<any JSON>"
    }
  ]
}
```

### `GET Produk/getLowestPriceOfThisProduct/{idProduk}`

App method: `HomeApi.getLowestPrice`.

Request:

```http
GET Produk/getLowestPriceOfThisProduct/{idProduk}
```

Response (`HomeResponses.LowestPriceResponse`):

```json
{
  "lowestPrice": {
    "diskon": 0,
    "harga": 0,
    "hargaBersih": 0
  }
}
```

## NotifikasiApi

### `GET User/getNotifikasi`

App method: `NotifikasiApi.getNotifikasi`.

Request:

```http
GET User/getNotifikasi
Authorization: Bearer <JWT>
```

Response (`NotifikasiResponses.InfoList`):

```json
{
  "notifikasi": [
    {
      "idNotifikasi": "string",
      "judul": "string",
      "pesan": "string",
      "foto": "string",
      "isRead": true,
      "createdAt": "string"
    }
  ]
}
```

### `GET User/getNotifikasiKhusus`

App method: `NotifikasiApi.getNotifikasiKhusus`.

Request:

```http
GET User/getNotifikasiKhusus
Authorization: Bearer <JWT>
```

Response (`NotifikasiResponses.KhususList`):

```json
{
  "notifikasi": [
    {
      "id": "string",
      "judul": "string",
      "isi": "string",
      "routeName": "string",
      "routeParams": "string",
      "isRead": true,
      "jenisNotifikasi": "string",
      "waktu": "string"
    }
  ]
}
```

### `GET User/getNotifikasiKhusus/{idUser}`

App method: `NotifikasiApi.getNotifikasiKhusus`.

Request:

```http
GET User/getNotifikasiKhusus/{idUser}
Authorization: Bearer <JWT>
```

Response (`NotifikasiResponses.KhususList`):

```json
{
  "notifikasi": [
    {
      "id": "string",
      "judul": "string",
      "isi": "string",
      "routeName": "string",
      "routeParams": "string",
      "isRead": true,
      "jenisNotifikasi": "string",
      "waktu": "string"
    }
  ]
}
```

### `GET User/getJumlahNotifikasi`

App method: `NotifikasiApi.getJumlahNotifikasi`.

Request:

```http
GET User/getJumlahNotifikasi?status=<status>
Authorization: Bearer <JWT>
```

Response (`NotifikasiResponses.Count`):

```json
{
  "jumlahNotifikasi": 0
}
```

### `GET User/getJumlahNotifikasiKhusus`

App method: `NotifikasiApi.getJumlahNotifikasiKhusus`.

Request:

```http
GET User/getJumlahNotifikasiKhusus?status=<status>
Authorization: Bearer <JWT>
```

Response (`NotifikasiResponses.Count`):

```json
{
  "jumlahNotifikasi": 0
}
```

### `GET Notifikasi/getNotifikasi/{idNotifikasi}`

App method: `NotifikasiApi.getDetailNotifikasi`.

Request:

```http
GET Notifikasi/getNotifikasi/{idNotifikasi}
Authorization: Bearer <JWT>
```

Response (`NotifikasiResponses.Detail`):

```json
{
  "notifikasi": {
    "idNotifikasi": "string",
    "judul": "string",
    "pesan": "string",
    "foto": "string",
    "isRead": true,
    "createdAt": "string"
  }
}
```

### `POST Notifikasi/setAs`

App method: `NotifikasiApi.setAs`.

Request:

```http
POST Notifikasi/setAs
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "idNotifikasiItem": "<idNotifikasiItem>",
  "setAs": "<setAs>",
  "jenisNotifikasi": "<jenisNotifikasi>"
}
```

Response (`NotifikasiResponses.SetAs`):

```json
{
  "statusSetAs": true
}
```

## PaketApi

### `GET Produk/detailPaket`

App method: `PaketApi.detailPaket`.

Request:

```http
GET Produk/detailPaket?idPaket=<idPaket>
```

Response (`PaketResponses.Detail`):

```json
{
  "detailPaket": {
    "id": "string",
    "nama": "string",
    "foto": "string",
    "harga": 0,
    "diskon": 0,
    "keterangan": "string"
  }
}
```

### `GET Produk/getPaketLainnya`

App method: `PaketApi.getPaketLainnya`.

Request:

```http
GET Produk/getPaketLainnya?idPaket=<idPaket>
```

Response (`PaketResponses.LainnyaList`):

```json
{
  "paketLainnya": [
    {
      "id": "string",
      "nama": "string",
      "foto": "string"
    }
  ]
}
```

### `POST Keranjang/saveCart`

App method: `PaketApi.saveCart`.

Legacy PaketModel.addToCart (jenisProduk=paket, quantity=1).

Request:

```http
POST Keranjang/saveCart
Content-Type: application/json

{
  "cartCode": "<cartCode>",
  "idProduk": "<idProduk>",
  "jenisProduk": "nonpaket",
  "quantity": "1",
  "jadwalDeliveryCode": "<jadwalDeliveryCode>",
  "dayDelivery": "Hari Pengantaran"
}
```

Response (`AddToCartResult`):

```json
{
  "statusAddCart": true,
  "message": "string"
}
```

## PaymentApi

### `GET Xendit/getPaymentChannels`

App method: `PaymentApi.getPaymentChannels`.

Request:

```http
GET Xendit/getPaymentChannels
Authorization: Bearer <JWT>
```

Response (`PaymentChannelsResponse`):

```json
{
  "paymentChannels": [
    {
      "channel_code": "string",
      "name": "string",
      "minimalPembayaran": 0
    }
  ]
}
```

### `POST Transaksi/checkout`

App method: `PaymentApi.checkout`.

Request:

```http
POST Transaksi/checkout
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "namaPenerima": "<namaPenerima>",
  "nomorHPPenerima": "<nomorHPPenerima>",
  "alamatDetailLengkap": "<alamatDetailLengkap>",
  "catatanTambahan": "<catatanTambahan>",
  "jadwalPengiriman": "<jadwalPengiriman>",
  "cartCode": "<cartCode>",
  "voucher": "0",
  "latDelivery": "<latDelivery>",
  "longDelivery": "<longDelivery>",
  "jadwalDeliveryCode": "<jadwalDeliveryCode>",
  "isPointUsed": "false",
  "diskonKategori": "<diskonKategori>",
  "kodePromo": "<kodePromo>",
  "metodePembayaran": "<metodePembayaran>",
  "namaMetodePembayaran": "<namaMetodePembayaran>",
  "moreDataMetodePembayaran": "<moreDataMetodePembayaran>",
  "alamatDelivery": "<alamatDelivery>"
}
```

Response (`CheckoutResponse`):

```json
{
  "statusCheckout": true,
  "message": "string",
  "idTransaksi": "string",
  "metodePembayaran": "string",
  "paymentResponse": "<any JSON>"
}
```

## ProdukApi

### `GET Produk/getProduk`

App method: `ProdukApi.getProduk`.

The product list, a page of 10 at a time (`useLimit` "y"), optionally filtered by name.

Request:

```http
GET Produk/getProduk?jadwalDeliveryCode=<jadwalDeliveryCode>&startFrom=<startFrom>&useLimit=<useLimit>&cariProduk=<cariProduk>
Authorization: Bearer <JWT>
```

Response (`ProdukResponses.ProdukList`):

```json
{
  "listProduk": [
    {
      "id": "string",
      "nama": "string",
      "foto": "string",
      "namaKategoriProduk": "string",
      "keterangan": "string",
      "keteranganSingkat": "string",
      "stokAvailable": true,
      "promo": [
        "<any JSON>"
      ]
    }
  ]
}
```

### `GET Produk/detailProduk`

App method: `ProdukApi.detailProduk`.

Request:

```http
GET Produk/detailProduk?idProduk=<idProduk>
```

Response (`ProdukResponses.Detail`):

```json
{
  "detailProduk": {
    "id": "string",
    "nama": "string",
    "foto": "string",
    "namaKategoriProduk": "string",
    "keterangan": "string",
    "keteranganSingkat": "string",
    "stokAvailable": true,
    "promo": [
      "<any JSON>"
    ]
  }
}
```

### `GET Produk/getLowestPriceOfThisProduct/{idProduk}`

App method: `ProdukApi.getLowestPrice`.

Request:

```http
GET Produk/getLowestPriceOfThisProduct/{idProduk}
```

Response (`ProdukResponses.LowestPriceResponse`):

```json
{
  "lowestPrice": {
    "harga": 0,
    "diskon": 0,
    "hargaBersih": 0
  }
}
```

### `GET Produk/getHargaProduk`

App method: `ProdukApi.getHargaProduk`.

Request:

```http
GET Produk/getHargaProduk?idProduk=<idProduk>
```

Response (`ProdukResponses.HargaList`):

```json
{
  "hargaProduk": [
    {
      "id": "string",
      "nama": "string",
      "harga": 0,
      "diskon": 0,
      "jenisHarga": "string",
      "isPromo": "<any JSON>",
      "idSatuanProduk": "string",
      "namaPromo": "string",
      "namaSatuan": "string",
      "maxOrder": "string",
      "hargaSatuan": 0,
      "diskonSatuan": 0,
      "diskonPromo": 0,
      "idPromo": "string"
    }
  ]
}
```

### `GET Produk/getProdukLainnya`

App method: `ProdukApi.getProdukLainnya`.

Request:

```http
GET Produk/getProdukLainnya?idProduk=<idProduk>
```

Response (`ProdukResponses.LainnyaList`):

```json
{
  "produkLainnya": [
    {
      "id": "string",
      "nama": "string",
      "foto": "string"
    }
  ]
}
```

### `GET Produk/isThisProductInCA`

App method: `ProdukApi.isThisProductInCA`.

Request:

```http
GET Produk/isThisProductInCA?idProduk=<idProduk>&latitude=<latitude>&longitude=<longitude>
```

Response (`ProdukResponses.InCa`):

```json
{
  "isThisProductInCA": true
}
```

## PromoApi

### `GET Promo/getPromoAroundOfThisLatLong`

App method: `PromoApi.getPromoAroundOfThisLatLong`.

Request:

```http
GET Promo/getPromoAroundOfThisLatLong?latitude=<latitude>&longitude=<longitude>&jadwalDeliveryCode=<jadwalDeliveryCode>
Authorization: Bearer <JWT>
```

Response (`PromoResponses.ProdukPromoList`):

```json
{
  "produkPromo": [
    {
      "idPromo": "string",
      "idPromoItem": "string",
      "idProduk": "string",
      "idHargaProduk": "string",
      "namaProduk": "string",
      "keteranganSingkat": "string",
      "fotoProduk": "string",
      "maxOrder": "string",
      "hargaProduk": "string",
      "diskonProduk": "string",
      "diskonPromo": "string"
    }
  ]
}
```

### `GET Promo/getPromoItem/{idPromoItem}`

App method: `PromoApi.getPromoItem`.

Request:

```http
GET Promo/getPromoItem/{idPromoItem}?select=<select>
```

Response (`PromoResponses.Detail`):

```json
{
  "promoItem": {
    "idPromo": "string",
    "statusPromo": "string",
    "catatan": "string",
    "diskon": "string",
    "detailPromo": {
      "nama": "string",
      "status": "string",
      "deskripsi": "string"
    },
    "detailProduk": {
      "id": "string",
      "nama": "string",
      "foto": "string",
      "namaKategoriProduk": "string",
      "keterangan": "string"
    },
    "detailSatuan": {
      "id": "string",
      "nama": "string",
      "harga": "string",
      "diskon": "string"
    }
  }
}
```

## PushApi

### `POST User/registerFcmToken`

App method: `PushApi.registerFcmToken`.

Links this device's FCM token to the signed-in user (idempotent: re-sending the same token is fine).

Request:

```http
POST User/registerFcmToken
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "fcmToken": "<fcmToken>",
  "platform": "<platform>"
}
```

Response (`PushResponses.Register`):

```json
{
  "statusRegister": true
}
```

### `POST User/unregisterFcmToken`

App method: `PushApi.unregisterFcmToken`.

Unlinks the token from the user (sign-out).

Request:

```http
POST User/unregisterFcmToken
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "fcmToken": "<fcmToken>"
}
```

Response (`PushResponses.Unregister`):

```json
{
  "statusUnregister": true
}
```

## RepeatOrderApi

### `GET RO/getItemTransaksi`

App method: `RepeatOrderApi.getItemTransaksi`.

Request:

```http
GET RO/getItemTransaksi?idTransaksi=<idTransaksi>&latDelivery=<latDelivery>&longDelivery=<longDelivery>&jadwalDeliveryCode=<jadwalDeliveryCode>
Authorization: Bearer <JWT>
```

Response (`RepeatOrderResponses.Items`):

```json
{
  "isAvailable": true,
  "itemTransaksi": [
    {
      "id": "string",
      "namaProduk": "string",
      "namaSatuan": "string",
      "jenisProduk": "string",
      "catatan": "string",
      "status": "string",
      "quantity": 0,
      "harga": 0,
      "diskon": 0,
      "diskonPromo": 0
    }
  ]
}
```

### `POST RO/addToCart`

App method: `RepeatOrderApi.addToCart`.

Request:

```http
POST RO/addToCart
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "idTransaksi": "<idTransaksi>",
  "cartCode": "<cartCode>"
}
```

Response (`RepeatOrderResponses.AddToCart`):

```json
{
  "statusAddCart": true,
  "message": "string"
}
```

## ResepApi

### `GET Resep/getResep`

App method: `ResepApi.getResep`.

Request:

```http
GET Resep/getResep
```

Response (`RecipeResponses.ResepList`):

```json
{
  "dataResep": [
    {
      "id": "string",
      "nama": "string",
      "foto": "string",
      "keterangan": "string",
      "igLink": "string",
      "fbLink": "string"
    }
  ]
}
```

### `GET Resep/getResep/{idResep}`

App method: `ResepApi.getDetailResep`.

Request:

```http
GET Resep/getResep/{idResep}?entitasResep=<entitasResep>
```

Response (`RecipeResponses.ResepDetail`):

```json
{
  "dataResep": {
    "id": "string",
    "nama": "string",
    "foto": "string",
    "keterangan": "string",
    "igLink": "string",
    "fbLink": "string"
  }
}
```

### `GET Resep/getResepItemFromThisResep/{idResep}`

App method: `ResepApi.getResepItem`.

Request:

```http
GET Resep/getResepItemFromThisResep/{idResep}?entitasResepItem=<entitasResepItem>
```

Response (`RecipeResponses.ItemList`):

```json
{
  "dataResepItem": [
    {
      "id": "string",
      "namaProduk": "string",
      "takaran": "string",
      "idProduk": "string"
    }
  ]
}
```

### `GET Produk/getHargaProduk`

App method: `ResepApi.getHargaProduk`.

Duplicate of the product slice's endpoint (units + prices of a product).

Request:

```http
GET Produk/getHargaProduk?idProduk=<idProduk>
```

Response (`RecipeResponses.SatuanList`):

```json
{
  "hargaProduk": [
    {
      "id": "string",
      "nama": "string",
      "harga": "string",
      "diskon": "string",
      "jenisHarga": "string"
    }
  ]
}
```

### `POST Keranjang/belanjaResep`

App method: `ResepApi.belanjaResep`.

Request:

```http
POST Keranjang/belanjaResep
Content-Type: application/json

{
  "cartCode": "<cartCode>",
  "selectedItems": {
    "<idItem>": {
      "idProduk": "<idProduk>",
      "quantity": "<quantity>",
      "idSatuan": "<idSatuan>",
      "subTotal": 0
    }
  },
  "dayDelivery": "Hari Pengantaran",
  "jadwalDeliveryCode": "<jadwalDeliveryCode>"
}
```

Response (`RecipeResponses.Belanja`):

```json
{
  "statusBelanjaResep": true
}
```

## SearchApi

### `GET Konten/pencarian`

App method: `SearchApi.pencarian`.

Request:

```http
GET Konten/pencarian?term=<term>
```

Response (`SearchResponses.Suggestions`):

```json
{
  "hasilPencarian": [
    {
      "id": "string",
      "jenisItem": "string",
      "fotoProduk": "string",
      "namaProduk": "string"
    }
  ]
}
```

## SettingApi

### `GET Pengaturan/getSettingValue/{key}`

App method: `SettingApi.getSettingValue`.

`{"settingValue": <any>`}.

Request:

```http
GET Pengaturan/getSettingValue/{key}
```

Response (`Map<String, Object>`):

```json
{
  "<key>": "<value>"
}
```

### `GET Pengaturan/getContactUsWhatsappConfiguration`

App method: `SettingApi.getContactUsWhatsappConfiguration`.

`{"contactUsWhatsappConfiguration": <object|false>`}.

Request:

```http
GET Pengaturan/getContactUsWhatsappConfiguration
```

Response (`Map<String, Object>`):

```json
{
  "<key>": "<value>"
}
```

## TransaksiApi

### `GET User/getRiwayatTransaksi`

App method: `TransaksiApi.getRiwayatTransaksi`.

Request:

```http
GET User/getRiwayatTransaksi?statusTransaksi=<statusTransaksi>&statusTransaksiHariIni=<statusTransaksiHariIni>&rentangAwal=<rentangAwal>&rentangAkhir=<rentangAkhir>
Authorization: Bearer <JWT>
```

Response (`TransaksiResponses.RiwayatList`):

```json
{
  "riwayatTransaksi": [
    {
      "id": "string",
      "cartCode": "string",
      "nomorTransaksi": "string",
      "status": "string",
      "statusReason": "string",
      "catatan": "string",
      "dayDelivery": "string",
      "jadwalDelivery": "string",
      "formattedTanggalDelivery": "string",
      "waktuCheckout": "string",
      "alamatDelivery": "string",
      "alamatDetailLengkap": "string",
      "metodePembayaran": "string",
      "namaMetodePembayaran": "string",
      "jenisVoucher": "string",
      "totalBelanja": 0,
      "ongkir": 0,
      "potonganWallet": 0,
      "nilaiVoucher": 0,
      "pointOut": 0,
      "diskonKategori": 0,
      "diskonPromo": 0,
      "moreDataMetodePembayaran": 0,
      "refundedOngkir": 0,
      "refundedUangBelanja": 0
    }
  ]
}
```

### `GET Transaksi/getTransaksi/{id}`

App method: `TransaksiApi.getTransaksi`.

Request:

```http
GET Transaksi/getTransaksi/{id}?entitasTransaksi=<entitasTransaksi>
Authorization: Bearer <JWT>
```

Response (`TransaksiResponses.Detail`):

```json
{
  "transaksi": {
    "id": "string",
    "cartCode": "string",
    "nomorTransaksi": "string",
    "status": "string",
    "statusReason": "string",
    "catatan": "string",
    "dayDelivery": "string",
    "jadwalDelivery": "string",
    "formattedTanggalDelivery": "string",
    "waktuCheckout": "string",
    "alamatDelivery": "string",
    "alamatDetailLengkap": "string",
    "metodePembayaran": "string",
    "namaMetodePembayaran": "string",
    "jenisVoucher": "string",
    "totalBelanja": 0,
    "ongkir": 0,
    "potonganWallet": 0,
    "nilaiVoucher": 0,
    "pointOut": 0,
    "diskonKategori": 0,
    "diskonPromo": 0,
    "moreDataMetodePembayaran": 0,
    "refundedOngkir": 0,
    "refundedUangBelanja": 0
  }
}
```

### `GET Transaksi/getItemTransaksiFromTransaksi/{id}`

App method: `TransaksiApi.getItemTransaksi`.

Request:

```http
GET Transaksi/getItemTransaksiFromTransaksi/{id}?entitasItemTransaksi=<entitasItemTransaksi>
Authorization: Bearer <JWT>
```

Response (`TransaksiResponses.Items`):

```json
{
  "itemTransaksi": [
    {
      "id": "string",
      "namaProduk": "string",
      "namaSatuan": "string",
      "jenisProduk": "string",
      "catatan": "string",
      "status": "string",
      "quantity": 0,
      "harga": 0,
      "diskon": 0,
      "diskonPromo": 0
    }
  ]
}
```

### `GET Transaksi/getJumlahItemDiantar/{id}`

App method: `TransaksiApi.getJumlahItemDiantar`.

Request:

```http
GET Transaksi/getJumlahItemDiantar/{id}
Authorization: Bearer <JWT>
```

Response (`TransaksiResponses.Diantar`):

```json
{
  "jumlahBarangDiantar": 0
}
```

### `POST Transaksi/pesananDiterima/{id}`

App method: `TransaksiApi.pesananDiterima`.

POST without a body.

Request:

```http
POST Transaksi/pesananDiterima/{id}
Authorization: Bearer <JWT>
```

Response (`TransaksiResponses.PesananDiterima`):

```json
{
  "pesananDiterima": true
}
```

### `POST Transaksi/belanjaUlang/{id}`

App method: `TransaksiApi.belanjaUlang`.

Request:

```http
POST Transaksi/belanjaUlang/{id}
Content-Type: application/json

{
  "dayDelivery": "Hari Pengantaran",
  "jadwalDeliveryCode": "<jadwalDeliveryCode>",
  "cartCode": "<cartCode>"
}
```

Response (`TransaksiResponses.BelanjaUlang`):

```json
{
  "statusBelanjaUlang": true,
  "message": "string"
}
```

### `GET gift/gift-transaksi`

App method: `TransaksiApi.getGiftTransaksi`.

Request:

```http
GET gift/gift-transaksi?cartCode=<cartCode>
```

Response (`TransaksiResponses.GiftList`):

```json
{
  "giftTransaksi": [
    {
      "id": "string",
      "idGift": "string",
      "cartCode": "string",
      "createdAt": "string"
    }
  ]
}
```

### `GET Gift/getGift/{idGift}`

App method: `TransaksiApi.getGift`.

Request:

```http
GET Gift/getGift/{idGift}
```

Response (`TransaksiResponses.GiftDetail`):

```json
{
  "gift": {
    "id": "string",
    "nama": "string",
    "deskripsi": "string"
  }
}
```

## UserApi

### `GET User/getDetailUser`

App method: `UserApi.getDetailUser`.

`getUserEntities`: optional comma separated column list (legacy "select").

Request:

```http
GET User/getDetailUser?getUserEntities=<select>
Authorization: Bearer <JWT>
```

Response (`AccountResponses.DetailUser`):

```json
{
  "detailUser": {
    "id": "string",
    "nama": "string",
    "username": "string",
    "alamat": "string",
    "nomorHP": "string",
    "foto": "string",
    "kodeReferal": "string",
    "createdAt": "string"
  }
}
```

### `GET User/getTotalDompet`

App method: `UserApi.getTotalDompet`.

Request:

```http
GET User/getTotalDompet
Authorization: Bearer <JWT>
```

Response (`AccountResponses.TotalDompet`):

```json
{
  "totalDompet": 0
}
```

### `GET User/getJumlahVoucher`

App method: `UserApi.getJumlahVoucher`.

Request:

```http
GET User/getJumlahVoucher
Authorization: Bearer <JWT>
```

Response (`AccountResponses.JumlahVoucher`):

```json
{
  "jumlahVoucher": 0
}
```

### `GET Point/getPoint`

App method: `UserApi.getPoint`.

Legacy getPoint("point=true").

Request:

```http
GET Point/getPoint?point=<point>
Authorization: Bearer <JWT>
```

Response (`AccountResponses.Point`):

```json
{
  "response": {
    "point": 0
  }
}
```

### `GET User/getRiwayatDompet`

App method: `UserApi.getRiwayatDompet`.

`jenisDompet`: "referral", "potongantransaksi" or null for all.

Request:

```http
GET User/getRiwayatDompet?jenisDompet=<jenisDompet>
Authorization: Bearer <JWT>
```

Response (`AccountResponses.RiwayatDompetList`):

```json
{
  "riwayatDompet": [
    {
      "jenis": "string",
      "keterangan": "string",
      "createdAt": "string",
      "nilai": 0
    }
  ]
}
```

### `GET User/getRiwayatPoint`

App method: `UserApi.getRiwayatPoint`.

Request:

```http
GET User/getRiwayatPoint
Authorization: Bearer <JWT>
```

Response (`AccountResponses.RiwayatPointList`):

```json
{
  "riwayatPoint": [
    {
      "keterangan": "string",
      "createdFrom": "string",
      "nilai": 0
    }
  ]
}
```

### `POST User/updateProfil`

App method: `UserApi.updateProfil`.

Request:

```http
POST User/updateProfil
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "namaLengkap": "<namaLengkap>",
  "alamat": "<alamat>",
  "username": "<username>"
}
```

Response (`AccountResponses.UpdateProfil`):

```json
{
  "statusUpdateProfil": true
}
```

### `POST User/updateFotoProfil`

App method: `UserApi.updateFotoProfil`.

Request:

```http
POST User/updateFotoProfil
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "fotoProfil": "<fotoProfil>"
}
```

Response (`AccountResponses.UpdateFoto`):

```json
{
  "saveFotoProfil": true
}
```

### `POST User/gantiPassword`

App method: `UserApi.gantiPassword`.

Request:

```http
POST User/gantiPassword
Authorization: Bearer <JWT>
Content-Type: application/json

{
  "passwordLama": "<passwordLama>",
  "passwordBaru": "<passwordBaru>",
  "konfirmasiPasswordBaru": "<konfirmasiPasswordBaru>"
}
```

Response (`AccountResponses.GantiPassword`):

```json
{
  "statusGantiPassword": true
}
```

## VoucherApi

### `GET User/getVoucher`

App method: `VoucherApi.getVoucher`.

Request:

```http
GET User/getVoucher?entitasVoucher=<entitasVoucher>
Authorization: Bearer <JWT>
```

Response (`VoucherResponses.VoucherList`):

```json
{
  "listVoucher": [
    {
      "id": "string",
      "foto": "string",
      "nama": "string",
      "jenis": "string",
      "nilai": "string",
      "kode": "string",
      "minimumBelanja": "string",
      "maximumUsage": "string",
      "usageCount": "string",
      "berlakuMulai": "string",
      "berlakuSampai": "string",
      "deskripsi": "string"
    }
  ]
}
```

### `GET Voucher/getVoucher/{idVoucher}`

App method: `VoucherApi.getVoucherDetail`.

Request:

```http
GET Voucher/getVoucher/{idVoucher}
Authorization: Bearer <JWT>
```

Response (`VoucherResponses.Detail`):

```json
{
  "dataVoucher": {
    "id": "string",
    "foto": "string",
    "nama": "string",
    "jenis": "string",
    "nilai": "string",
    "kode": "string",
    "minimumBelanja": "string",
    "maximumUsage": "string",
    "usageCount": "string",
    "berlakuMulai": "string",
    "berlakuSampai": "string",
    "deskripsi": "string"
  }
}
```
