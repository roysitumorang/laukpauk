# Old endpoints used by laukpauk_mobile

Every backend call the Android app ([roysitumorang/laukpauk_mobile](https://github.com/roysitumorang/laukpauk_mobile), branch `dev`) makes, taken from its Retrofit interfaces in `android/app/src/main/java/id/laukpauk/data/remote/`. The list reflects the app after its seller and market code was removed (all products belong to a single seller) and after every call moved to JSON bodies, JWT auth and the `X-Version-Code` / `X-Request-Id` headers ([laukpauk_mobile#12](https://github.com/roysitumorang/laukpauk_mobile/pull/12)). Sign-in is `POST users/sign_in` ([laukpauk_mobile#13](https://github.com/roysitumorang/laukpauk_mobile/pull/13)), buyer registration is `POST users` ([laukpauk_mobile#14](https://github.com/roysitumorang/laukpauk_mobile/pull/14)), the home banners come from `GET buyer/banners` ([laukpauk_mobile#21](https://github.com/roysitumorang/laukpauk_mobile/pull/21)) and the product lists and product search use `GET buyer/products` ([laukpauk_mobile#22](https://github.com/roysitumorang/laukpauk_mobile/pull/22)). These are the endpoints the app calls, not the endpoints this repository serves; use this list to map the app onto the current API.

## Conventions

- **Base URL:** paths are relative to `API_BASE_URL` (a build setting, e.g. `https://<host>/api/`).
- **Headers on every backend request** (the examples below leave the first two out):
  - `X-Version-Code`: the app's version code (build number).
  - `X-Request-Id`: a random UUID v4, new for each request.
  - `Accept: application/json`.
- **Auth:** JWT. While a user is signed in, every backend request carries `Authorization: Bearer <JWT>` (the `id_token` returned by `POST users/sign_in`, until its `expired_at` passes). The examples show the header on the calls that need a signed-in user; no call sends the token in the URL or the body.
- **Bodies:** every `POST` sends `Content-Type: application/json` with a JSON object, except `Transaksi/pesananDiterima/{id}`, which is sent with no body. Keys whose value is null are left out. Most values are JSON strings, because the app holds them as text. The profile photo is sent as a base64 string.
- **Path parameters** are shown as `{name}`; placeholder values as `<name>`.
- **Errors:** a failed call answers a non-2xx status; the app takes the reason from the `message` of the JSON error body when there is one.
- **Responses** are JSON. The examples list only the fields the app reads (from its response classes), with type placeholders: `"string"`, `0`, `0.0`, `true`.
- **Overloads:** where the app calls a path both with and without a trailing `/{id}`, both forms are listed.

## Contents

- [AccountApi](#accountapi) (5)
- [AuthApi](#authapi) (7)
- [CartApi](#cartapi) (10)
- [CheckoutApi](#checkoutapi) (14)
- [DeliveryApi](#deliveryapi) (5)
- [HomeApi](#homeapi) (9)
- [NotificationApi](#notificationapi) (7)
- [PackApi](#packapi) (3)
- [PaymentApi](#paymentapi) (2)
- [ProductApi](#productapi) (7)
- [PromoApi](#promoapi) (2)
- [PushApi](#pushapi) (2)
- [RepeatOrderApi](#repeatorderapi) (2)
- [RecipeApi](#recipeapi) (5)
- [SettingApi](#settingapi) (2)
- [TransactionApi](#transactionapi) (8)
- [UserApi](#userapi) (9)
- [VoucherApi](#voucherapi) (2)

101 endpoint definitions in total.

## Index

| Method | Path | JWT required | App method |
|---|---|---|---|
| GET | `Auth/removeToken` | yes | `AccountApi.removeToken` |
| POST | `Auth/gantiNomor` | yes | `AccountApi.changeNumber` |
| POST | `Auth/verifikasiNomor` | yes | `AccountApi.numberVerification` |
| GET | `Auth/jumlahGantiNomorHPForToday` | yes | `AccountApi.phoneNumberChangeCountForToday` |
| GET | `Pengaturan/getReferralConfiguration` | yes | `AccountApi.getReferralConfiguration` |
| POST | `users/sign_in` | no | `AuthApi.signIn` |
| POST | `users` | no | `AuthApi.signUp` |
| POST | `Auth/konfirmasiOTP` | no | `AuthApi.confirmOtp` |
| POST | `Auth/verifikasiUser` | no | `AuthApi.userVerification` |
| POST | `Auth/resendOTP` | yes | `AuthApi.resendOTP` |
| POST | `Auth/lupaPassword` | no | `AuthApi.forgotPassword` |
| GET | `User/getDetailUser` | yes | `AuthApi.getDetailUser` |
| GET | `User/getJumlahCart` | no | `CartApi.getCartQuantity` |
| GET | `User/getCart` | no | `CartApi.getCart` |
| GET | `User/getTotalKeranjang` | no | `CartApi.getCartTotal` |
| GET | `Keranjang/bersihkanKeranjang` | no | `CartApi.clearCart` |
| POST | `Keranjang/saveCart` | no | `CartApi.saveCart` |
| POST | `Keranjang/saveCart/{idItem}` | no | `CartApi.updateCart` |
| POST | `Keranjang/deleteCart/{idCart}` | no | `CartApi.deleteCart` |
| GET | `Pengaturan/getWalletConfiguration` | no | `CartApi.getWalletConfiguration` |
| GET | `gift/gift-transaksi` | no | `CartApi.getTransactionGift` |
| GET | `Gift/getGift/{idGift}` | no | `CartApi.getGift` |
| GET | `Auth/cekVerifikasiUser` | yes | `CheckoutApi.checkUserVerification` |
| GET | `User/getDetailUser` | yes | `CheckoutApi.getDetailUser` |
| POST | `JadwalDelivery/isJadwalDeliveryOpened` | no | `CheckoutApi.isDeliveryScheduleOpened` |
| GET | `User/getTotalKeranjang` | no | `CheckoutApi.getCartTotal` |
| GET | `User/getTotalOngkir` | no | `CheckoutApi.getTotalShippingFee` |
| GET | `User/getJumlahCart` | no | `CheckoutApi.getCartQuantity` |
| GET | `Transaksi/getJenisOngkir` | no | `CheckoutApi.getShippingFeeType` |
| GET | `User/getDiskonKategori` | no | `CheckoutApi.getCategoryDiscount` |
| GET | `Pengaturan/getWalletConfiguration` | no | `CheckoutApi.getWalletConfiguration` |
| GET | `User/getTotalDompet` | yes | `CheckoutApi.getWalletTotal` |
| GET | `Point/getPoint` | yes | `CheckoutApi.getPoint` |
| GET | `User/getVoucher` | yes | `CheckoutApi.getVoucher` |
| GET | `Voucher/getStatusVoucher/{id}` | yes | `CheckoutApi.getStatusVoucher` |
| GET | `Promo/cekKodePromo/{kode}` | yes | `CheckoutApi.checkPromoCode` |
| GET | `JadwalDelivery/getAvailableJadwalDelivery` | no | `DeliveryApi.getAvailableDeliverySchedule` |
| GET | `Gift/getGift` | no | `DeliveryApi.getGifts` |
| GET | `Keranjang/bersihkanKeranjang` | no | `DeliveryApi.clearCart` |
| GET | `https://maps.googleapis.com/maps/api/geocode/json` | no | `DeliveryApi.reverseGeocode` |
| GET | `https://maps.googleapis.com/maps/api/place/textsearch/json` | no | `DeliveryApi.searchPlaces` |
| GET | `buyer/banners` | no | `HomeApi.getBanner` |
| GET | `User/getJumlahVoucher` | yes | `HomeApi.getVoucherCount` |
| GET | `User/getTotalDompet` | yes | `HomeApi.getWalletTotal` |
| GET | `Point/getPoint` | yes | `HomeApi.getPoint` |
| GET | `Konten/getKonten` | no | `HomeApi.getContent` |
| GET | `Konten/getKontenItemFromThisKonten/{idKonten}` | no | `HomeApi.getContentItem` |
| GET | `KategoriProduk/getKategoriProduk` | no | `HomeApi.getProductCategory` |
| GET | `Produk/getProductsInThisCategory/{idKategoriProduk}` | no | `HomeApi.getProductInThisCategory` |
| GET | `Produk/getLowestPriceOfThisProduct/{idProduk}` | no | `HomeApi.getLowestPrice` |
| GET | `User/getNotifikasi` | yes | `NotificationApi.getNotification` |
| GET | `User/getNotifikasiKhusus` | yes | `NotificationApi.getSpecialNotification` |
| GET | `User/getNotifikasiKhusus/{idUser}` | yes | `NotificationApi.getSpecialNotification` |
| GET | `User/getJumlahNotifikasi` | yes | `NotificationApi.getNotificationCount` |
| GET | `User/getJumlahNotifikasiKhusus` | yes | `NotificationApi.getSpecialNotificationCount` |
| GET | `Notifikasi/getNotifikasi/{idNotifikasi}` | yes | `NotificationApi.getNotificationDetail` |
| POST | `Notifikasi/setAs` | yes | `NotificationApi.setAs` |
| GET | `Produk/detailPaket` | no | `PackApi.packDetail` |
| GET | `Produk/getPaketLainnya` | no | `PackApi.getOtherPacks` |
| POST | `Keranjang/saveCart` | no | `PackApi.saveCart` |
| GET | `Xendit/getPaymentChannels` | yes | `PaymentApi.getPaymentChannels` |
| POST | `Transaksi/checkout` | yes | `PaymentApi.checkout` |
| GET | `buyer/products` | no | `ProductApi.buyerProducts` |
| GET | `data.next` of `buyer/products` | no | `ProductApi.buyerProductsPage` |
| GET | `Produk/detailProduk` | no | `ProductApi.productDetail` |
| GET | `Produk/getLowestPriceOfThisProduct/{idProduk}` | no | `ProductApi.getLowestPrice` |
| GET | `Produk/getHargaProduk` | no | `ProductApi.getProductPrice` |
| GET | `Produk/getProdukLainnya` | no | `ProductApi.getOtherProducts` |
| GET | `Produk/isThisProductInCA` | no | `ProductApi.isThisProductInCA` |
| GET | `Promo/getPromoAroundOfThisLatLong` | yes | `PromoApi.getPromoAroundOfThisLatLong` |
| GET | `Promo/getPromoItem/{idPromoItem}` | no | `PromoApi.getPromoItem` |
| POST | `User/registerFcmToken` | yes | `PushApi.registerFcmToken` |
| POST | `User/unregisterFcmToken` | yes | `PushApi.unregisterFcmToken` |
| GET | `RO/getItemTransaksi` | yes | `RepeatOrderApi.getTransactionItem` |
| POST | `RO/addToCart` | yes | `RepeatOrderApi.addToCart` |
| GET | `Resep/getResep` | no | `RecipeApi.getRecipe` |
| GET | `Resep/getResep/{idResep}` | no | `RecipeApi.getRecipeDetail` |
| GET | `Resep/getResepItemFromThisResep/{idResep}` | no | `RecipeApi.getRecipeItem` |
| GET | `Produk/getHargaProduk` | no | `RecipeApi.getProductPrice` |
| POST | `Keranjang/belanjaResep` | no | `RecipeApi.shopRecipe` |
| GET | `Pengaturan/getSettingValue/{key}` | no | `SettingApi.getSettingValue` |
| GET | `Pengaturan/getContactUsWhatsappConfiguration` | no | `SettingApi.getContactUsWhatsappConfiguration` |
| GET | `User/getRiwayatTransaksi` | yes | `TransactionApi.getTransactionRecords` |
| GET | `Transaksi/getTransaksi/{id}` | yes | `TransactionApi.getTransaction` |
| GET | `Transaksi/getItemTransaksiFromTransaksi/{id}` | yes | `TransactionApi.getTransactionItem` |
| GET | `Transaksi/getJumlahItemDiantar/{id}` | yes | `TransactionApi.getDeliveredItemCount` |
| POST | `Transaksi/pesananDiterima/{id}` | yes | `TransactionApi.orderReceived` |
| POST | `Transaksi/belanjaUlang/{id}` | no | `TransactionApi.reorder` |
| GET | `gift/gift-transaksi` | no | `TransactionApi.getTransactionGift` |
| GET | `Gift/getGift/{idGift}` | no | `TransactionApi.getGift` |
| GET | `User/getDetailUser` | yes | `UserApi.getDetailUser` |
| GET | `User/getTotalDompet` | yes | `UserApi.getWalletTotal` |
| GET | `User/getJumlahVoucher` | yes | `UserApi.getVoucherCount` |
| GET | `Point/getPoint` | yes | `UserApi.getPoint` |
| GET | `User/getRiwayatDompet` | yes | `UserApi.getWalletRecords` |
| GET | `User/getRiwayatPoint` | yes | `UserApi.getPointRecords` |
| POST | `User/updateProfil` | yes | `UserApi.updateProfile` |
| POST | `User/updateFotoProfil` | yes | `UserApi.updateProfilePhoto` |
| POST | `User/gantiPassword` | yes | `UserApi.changePassword` |
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

App method: `AccountApi.changeNumber`.

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

App method: `AccountApi.numberVerification`.

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

App method: `AccountApi.phoneNumberChangeCountForToday`.

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

The app reads the three attributes from `data` (the backend's response envelope) when it is present, and from the top level otherwise. It stores `id_token` as the session JWT, `expired_at` (RFC 3339 date-time) as its expiry, and `user` as the current user. From `user` it reads `id` (the user id) and `mobile_phone_confirmed_at` (null opens the phone-verification screen); the other `user` fields shown are the ones the backend's user model sends, and are stored as they come. After `expired_at` the app treats the session as signed out and stops sending the token. A 4xx status means the login or password is wrong. In the request, `password` is always the standard base64 (with padding) of the password's UTF-8 bytes, for example `cmFoYXNpYTEyMw==` for `rahasia123`; registration (`POST users`) encodes its password the same way, while `User/gantiPassword` still sends its passwords as typed.

### `POST users`

App method: `AuthApi.signUp`.

Buyer registration. Body: `name`, `username`, `mobile_phone`, `password` (base64 of the UTF-8 password) and `referral_code` (all strings). Answers the new user, shaped like `user` of sign-in, either at the top level or inside the backend's `data` envelope. A 4xx status carries the reason in `message`.

Request:

```http
POST users
Content-Type: application/json

{
  "name": "<name>",
  "username": "<username>",
  "mobile_phone": "<phone number>",
  "password": "<base64 of the password>",
  "referral_code": "<referral code>"
}
```

Response (the new user, shaped like `user` of sign-in):

```json
{
  "data": {
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
```

The app reads the user from `data` (the backend's response envelope) when it is present, and from the top level otherwise. It keeps `id` for the OTP confirmation that follows (`Auth/konfirmasiOTP`); the user is not signed in yet, and a new user's `mobile_phone_confirmed_at` is null. In the request, `password` is the standard base64 of the password's UTF-8 bytes, `referral_code` is left out when empty, and the password confirmation is checked in the app, not sent. A 4xx status means the registration was refused; the app shows the `message` of the error body.

### `POST Auth/konfirmasiOTP`

App method: `AuthApi.confirmOtp`.

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

App method: `AuthApi.userVerification`.

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

App method: `AuthApi.forgotPassword`.

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

App method: `CartApi.getCartQuantity`.

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

App method: `CartApi.getCartTotal`.

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

App method: `CartApi.clearCart`.

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

App method: `CartApi.getTransactionGift`.

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

App method: `CheckoutApi.checkUserVerification`.

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

App method: `CheckoutApi.isDeliveryScheduleOpened`.

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

App method: `CheckoutApi.getCartTotal`.

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

App method: `CheckoutApi.getTotalShippingFee`.

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

App method: `CheckoutApi.getCartQuantity`.

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

App method: `CheckoutApi.getShippingFeeType`.

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

App method: `CheckoutApi.getCategoryDiscount`.

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

App method: `CheckoutApi.getWalletTotal`.

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

App method: `CheckoutApi.checkPromoCode`.

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

App method: `DeliveryApi.getAvailableDeliverySchedule`.

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

App method: `DeliveryApi.clearCart`.

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

### `GET buyer/banners`

App method: `HomeApi.getBanner`.

Home page hero banners. The list is in the backend's `data` envelope; each banner has `id`, `name`, `description`, `photo` (a full image URL) and `visible`. The app skips banners with `visible: false`, shows `name` / `description` only when not empty; an `http://` photo is still loaded over `https://`. The other fields are ignored.

Request:

```http
GET buyer/banners
```

Response (`HomeResponses.Banners`):

```json
{
  "request_id": "19ecbcb8-2456-4f56-996b-b1ae2cb74abd",
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": [
    {
      "id": "01a1f1aa-ca52-7972-b7dc-d9ea83a10903",
      "record_status": 0,
      "name": "",
      "description": "",
      "photo": "https://sadia.id/images/01a1f1aa-ca1f-71b6-9944-975656e8022d.webp",
      "visible": true,
      "created_by": "01a1f1a9-4f55-70b9-b339-24501657e0cc",
      "created_at": "2021-11-19T16:07:41+07:00",
      "updated_by": "01a1f1a9-4f55-70b9-b339-24501657e0cc",
      "updated_at": "2021-11-19T16:07:41+07:00",
      "deleted_by": null,
      "deleted_at": null,
      "deletion_reason": null
    }
  ],
  "app": "laukpauk"
}
```

### `GET User/getJumlahVoucher`

App method: `HomeApi.getVoucherCount`.

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

App method: `HomeApi.getWalletTotal`.

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

### `GET Konten/getKonten`

App method: `HomeApi.getContent`.

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

App method: `HomeApi.getContentItem`.

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

App method: `HomeApi.getProductCategory`.

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

App method: `HomeApi.getProductInThisCategory`.

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

## NotificationApi

### `GET User/getNotifikasi`

App method: `NotificationApi.getNotification`.

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

App method: `NotificationApi.getSpecialNotification`.

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

App method: `NotificationApi.getSpecialNotification`.

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

App method: `NotificationApi.getNotificationCount`.

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

App method: `NotificationApi.getSpecialNotificationCount`.

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

App method: `NotificationApi.getNotificationDetail`.

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

App method: `NotificationApi.setAs`.

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

## PackApi

### `GET Produk/detailPaket`

App method: `PackApi.packDetail`.

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

App method: `PackApi.getOtherPacks`.

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

App method: `PackApi.saveCart`.

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

## ProductApi

### `GET buyer/products`

App method: `ProductApi.buyerProducts`.

Products for the home page "Produk" block, the product grid ("Lihat semua") and the product search (the grid's search box and the header search screen). Query: `limit` (products per page, the app sends 10) and `q` (search text, only when searching). The page is in the backend's `data` envelope: `rows`, `next` (full URL of the next page, `null` on the last page), `current` and `limit`. The app uses `id`, `name`, `photo` (full URL; `http://` is loaded over `https://`), `category_name`, `description`, `ready_stock`, `product_type` (1 = paket, opens the pack detail) and `lowest_selling_price` (the tile price; 0 = no price). The other fields are ignored.

Request:

```http
GET buyer/products?limit=10&q=<q>
```

Response (`ProductResponses.BuyerProducts`):

```json
{
  "request_id": "86609e54-c094-4ce5-aaf7-0a1b0d14c4d0",
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "current": "https://sadia.id/v1/buyer/products?limit=10",
    "next": "https://sadia.id/v1/buyer/products?limit=10&next_row_no=11",
    "limit": 10,
    "rows": [
      {
        "id": "01a1f1aa-c5ff-73dc-9618-01b224bed65f",
        "record_status": 0,
        "category_id": "01a1f1a9-ec59-7f0c-ab72-6be87c24cc24",
        "category_name": "Paket",
        "product_type": 1,
        "product_type_name": "paket",
        "name": "Belanja hemat 2 Hari Cuma 30 Ribuan!",
        "photo": "https://sadia.id/images/bundle-default.png",
        "description": "string",
        "ready_stock": true,
        "pinned": false,
        "hits": 0,
        "lowest_selling_price": 37000,
        "created_by": "01a1f1a9-4f55-70b9-b339-24501657e0cc",
        "created_at": "2021-11-25T10:15:48+07:00",
        "updated_by": "01a1f1a9-4f55-70b9-b339-24501657e0cc",
        "updated_at": "2021-11-25T10:15:48+07:00",
        "deleted_by": null,
        "deleted_at": null,
        "deletion_reason": null,
        "variants": []
      }
    ]
  },
  "app": "laukpauk"
}
```

### `GET` `data.next` of `buyer/products`

App method: `ProductApi.buyerProductsPage`.

The next page of a product list or search: the app requests the full URL in `data.next` as is when the list is scrolled to its end, and stops when `data.next` is `null`. Same response as `GET buyer/products`.

Request:

```http
GET https://sadia.id/v1/buyer/products?limit=10&next_row_no=11
```

### `GET Produk/detailProduk`

App method: `ProductApi.productDetail`.

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

App method: `ProductApi.getLowestPrice`.

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

App method: `ProductApi.getProductPrice`.

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

App method: `ProductApi.getOtherProducts`.

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

App method: `ProductApi.isThisProductInCA`.

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

App method: `RepeatOrderApi.getTransactionItem`.

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

## RecipeApi

### `GET Resep/getResep`

App method: `RecipeApi.getRecipe`.

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

App method: `RecipeApi.getRecipeDetail`.

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

App method: `RecipeApi.getRecipeItem`.

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

App method: `RecipeApi.getProductPrice`.

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

App method: `RecipeApi.shopRecipe`.

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

## TransactionApi

### `GET User/getRiwayatTransaksi`

App method: `TransactionApi.getTransactionRecords`.

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

App method: `TransactionApi.getTransaction`.

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

App method: `TransactionApi.getTransactionItem`.

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

App method: `TransactionApi.getDeliveredItemCount`.

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

App method: `TransactionApi.orderReceived`.

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

App method: `TransactionApi.reorder`.

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

App method: `TransactionApi.getTransactionGift`.

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

App method: `TransactionApi.getGift`.

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

App method: `UserApi.getWalletTotal`.

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

App method: `UserApi.getVoucherCount`.

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

App method: `UserApi.getWalletRecords`.

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

App method: `UserApi.getPointRecords`.

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

App method: `UserApi.updateProfile`.

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

App method: `UserApi.updateProfilePhoto`.

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

App method: `UserApi.changePassword`.

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
