# Old endpoints used by laukpauk_mobile

Every backend call the Android app ([roysitumorang/laukpauk_mobile](https://github.com/roysitumorang/laukpauk_mobile), branch `dev`) makes, taken from its Retrofit interfaces in `android/app/src/main/java/id/laukpauk/data/remote/`. The list reflects the app after its seller and market code was removed (all products belong to a single seller) and after every call moved to JSON bodies, JWT auth and the `X-Version-Code` / `X-Request-Id` headers ([laukpauk_mobile#12](https://github.com/roysitumorang/laukpauk_mobile/pull/12)). Sign-in is `POST buyer/users/sign_in` and buyer registration is `POST buyer/users` ([laukpauk_mobile#13](https://github.com/roysitumorang/laukpauk_mobile/pull/13), [#14](https://github.com/roysitumorang/laukpauk_mobile/pull/14), [#26](https://github.com/roysitumorang/laukpauk_mobile/pull/26)), the home banners come from `GET buyer/banners` ([laukpauk_mobile#21](https://github.com/roysitumorang/laukpauk_mobile/pull/21)), the product lists and product search use `GET buyer/products` ([laukpauk_mobile#22](https://github.com/roysitumorang/laukpauk_mobile/pull/22)), the product detail screen uses `GET buyer/products/{id}` ([laukpauk_mobile#23](https://github.com/roysitumorang/laukpauk_mobile/pull/23)) and the cart uses `GET` / `POST buyer/cart` ([laukpauk_mobile#25](https://github.com/roysitumorang/laukpauk_mobile/pull/25), [#27](https://github.com/roysitumorang/laukpauk_mobile/pull/27)), checkout's delivery address comes from `buyer/delivery_addresses` ([laukpauk_mobile#28](https://github.com/roysitumorang/laukpauk_mobile/pull/28), [#38](https://github.com/roysitumorang/laukpauk_mobile/pull/38)) and its delivery schedule options from `GET buyer/delivery_plans/now` ([laukpauk_mobile#29](https://github.com/roysitumorang/laukpauk_mobile/pull/29)), and its delivery cost is the address's `distance_km` times `delivery_cost_per_km` of `GET buyer/setting` ([laukpauk_mobile#39](https://github.com/roysitumorang/laukpauk_mobile/pull/39)). These are the endpoints the app calls, not the endpoints this repository serves; use this list to map the app onto the current API.

## Conventions

- **Base URL:** paths are relative to `API_BASE_URL` (a build setting, e.g. `https://<host>/api/`).
- **Headers on every backend request** (the examples below leave the first two out):
  - `X-Version-Code`: the app's version code (build number).
  - `X-Request-Id`: a random UUID v4, new for each request.
  - `Accept: application/json`.
- **Auth:** JWT. While a user is signed in, every backend request carries `Authorization: Bearer <JWT>` (the `id_token` returned by `POST buyer/users/sign_in`, until its `expired_at` passes). The examples show the header on the calls that need a signed-in user; no call sends the token in the URL or the body.
- **Bodies:** every `POST` sends `Content-Type: application/json` with a JSON object, except `Transaksi/pesananDiterima/{id}`, which is sent with no body. Keys whose value is null are left out. Most values are JSON strings, because the app holds them as text. The profile photo is sent as a base64 string.
- **Path parameters** are shown as `{name}`; placeholder values as `<name>`.
- **Errors:** a failed call answers a non-2xx status; the app takes the reason from the `message` of the JSON error body when there is one.
- **Responses** are JSON. The examples list only the fields the app reads (from its response classes), with type placeholders: `"string"`, `0`, `0.0`, `true`.
- **Overloads:** where the app calls a path both with and without a trailing `/{id}`, both forms are listed.

## Contents

- [AccountApi](#accountapi) (5)
- [AddressApi](#addressapi) (4)
- [AuthApi](#authapi) (5)
- [CartApi](#cartapi) (4)
- [CheckoutApi](#checkoutapi) (2)
- [DeliveryApi](#deliveryapi) (4)
- [HomeApi](#homeapi) (6)
- [NotificationApi](#notificationapi) (7)
- [PackApi](#packapi) (3)
- [PaymentApi](#paymentapi) (2)
- [ProductApi](#productapi) (8)
- [PromoApi](#promoapi) (2)
- [PushApi](#pushapi) (2)
- [RepeatOrderApi](#repeatorderapi) (2)
- [RecipeApi](#recipeapi) (3)
- [SettingApi](#settingapi) (3)
- [TransactionApi](#transactionapi) (7)
- [UserApi](#userapi) (6)
- [VoucherApi](#voucherapi) (2)

77 endpoint definitions in total.

## Index

| Method | Path | JWT required | App method |
|---|---|---|---|
| GET | `Auth/removeToken` | yes | `AccountApi.removeToken` |
| POST | `Auth/gantiNomor` | yes | `AccountApi.changeNumber` |
| POST | `Auth/verifikasiNomor` | yes | `AccountApi.numberVerification` |
| GET | `Auth/jumlahGantiNomorHPForToday` | yes | `AccountApi.phoneNumberChangeCountForToday` |
| GET | `Pengaturan/getReferralConfiguration` | yes | `AccountApi.getReferralConfiguration` |
| GET | `buyer/delivery_addresses` | yes | `AddressApi.list` |
| GET | `data.links.next` of `buyer/delivery_addresses` | yes | `AddressApi.page` |
| POST | `buyer/delivery_addresses` | yes | `AddressApi.create` |
| PUT | `buyer/delivery_addresses/{id}` | yes | `AddressApi.update` |
| POST | `buyer/users/sign_in` | no | `AuthApi.signIn` |
| POST | `buyer/users` | no | `AuthApi.signUp` |
| POST | `buyer/users/mobile_phone/confirm` | yes | `AuthApi.confirmMobilePhone` |
| POST | `Auth/resendOTP` | yes | `AuthApi.resendOTP` |
| POST | `Auth/lupaPassword` | no | `AuthApi.forgotPassword` |
| GET | `buyer/cart` | yes | `CartApi.buyerCart` |
| POST | `buyer/cart` | yes | `CartApi.saveBuyerCart` |
| GET | `Keranjang/bersihkanKeranjang` | no | `CartApi.clearCart` |
| POST | `Keranjang/saveCart` | no | `CartApi.saveCart` |
| POST | `JadwalDelivery/isJadwalDeliveryOpened` | no | `CheckoutApi.isDeliveryScheduleOpened` |
| GET | `buyer/promotions/{promo_code}` | yes | `CheckoutApi.buyerPromotion` |
| GET | `buyer/delivery_plans/now` | yes | `DeliveryApi.deliveryPlanNow` |
| GET | `Keranjang/bersihkanKeranjang` | no | `DeliveryApi.clearCart` |
| GET | `https://maps.googleapis.com/maps/api/geocode/json` | no | `DeliveryApi.reverseGeocode` |
| GET | `https://maps.googleapis.com/maps/api/place/textsearch/json` | no | `DeliveryApi.searchPlaces` |
| GET | `buyer/banners` | no | `HomeApi.getBanner` |
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
| GET | `buyer/payment_channels` | yes | `PaymentApi.buyerPaymentChannels` |
| POST | `buyer/transactions` | yes | `PaymentApi.createTransaction` |
| GET | `buyer/products` | no | `ProductApi.buyerProducts` |
| GET | `data.next` of `buyer/products` | no | `ProductApi.buyerProductsPage` |
| GET | `buyer/products/{id}` | no | `ProductApi.buyerProduct` |
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
| GET | `buyer/recipes` | no | `RecipeApi.buyerRecipes` |
| GET | `data.links.next` of `buyer/recipes` | no | `RecipeApi.buyerRecipesPage` |
| GET | `buyer/recipes/{id}` | no | `RecipeApi.buyerRecipe` |
| GET | `buyer/setting` | no | `SettingApi.buyerSetting` |
| GET | `Pengaturan/getSettingValue/{key}` | no | `SettingApi.getSettingValue` |
| GET | `Pengaturan/getContactUsWhatsappConfiguration` | no | `SettingApi.getContactUsWhatsappConfiguration` |
| GET | `buyer/transactions` | yes | `TransactionApi.buyerTransactions` |
| GET | `data.links.next` of `buyer/transactions` | yes | `TransactionApi.buyerTransactionsPage` |
| GET | `buyer/transactions/{id}` | yes | `TransactionApi.buyerTransaction` |
| GET | `Transaksi/getTransaksi/{id}` | yes | `TransactionApi.getTransaction` |
| GET | `Transaksi/getJumlahItemDiantar/{id}` | yes | `TransactionApi.getDeliveredItemCount` |
| POST | `Transaksi/pesananDiterima/{id}` | yes | `TransactionApi.orderReceived` |
| POST | `Transaksi/belanjaUlang/{id}` | no | `TransactionApi.reorder` |
| GET | `buyer/users/me` | yes | `UserApi.me` |
| GET | `User/getRiwayatDompet` | yes | `UserApi.getWalletRecords` |
| GET | `User/getRiwayatPoint` | yes | `UserApi.getPointRecords` |
| POST | `User/updateProfil` | yes | `UserApi.updateProfile` |
| POST | `User/updateFotoProfil` | yes | `UserApi.updateProfilePhoto` |
| POST | `User/gantiPassword` | yes | `UserApi.changePassword` |
| GET | `buyer/vouchers` | yes | `VoucherApi.buyerVouchers` |
| GET | `buyer/vouchers/{id}` | yes | `VoucherApi.buyerVoucher` |

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

## AddressApi

### `GET buyer/delivery_addresses`

App method: `AddressApi.list`.

The signed-in buyer's delivery addresses, listed when the buyer taps Checkout. The app asks `limit=100`; the backend always lists the main address first. Page-based: `data.info` and `data.links` (`next` is empty on the last page); the addresses are `data.rows`. The app shows `label` (+ "UTAMA" when `main`), `street_address` and `street_address_note`; choosing one sets the delivery latitude and longitude (`destination_latitude`, `destination_longitude`) and `street_address`. `distance_km` is read but not shown. An empty list goes straight to adding an address.

Request:

```http
GET buyer/delivery_addresses?limit=100
Authorization: Bearer <JWT>
```

Response (`AddressResponses.AddressList`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "links": {
      "first": "https://sadia.id/v1/buyer/delivery_addresses?limit=100&page=1",
      "previous": "",
      "current": "https://sadia.id/v1/buyer/delivery_addresses?limit=100&page=1",
      "next": "",
      "last": "https://sadia.id/v1/buyer/delivery_addresses?limit=100&page=1"
    },
    "info": {
      "limit": 100,
      "page": 1,
      "pages": 1,
      "total": 1
    },
    "rows": [
      {
        "id": "string",
        "record_status": 0,
        "user_id": "string",
        "origin_longitude": "98.6600",
        "origin_latitude": "3.5800",
        "destination_longitude": "98.6722",
        "destination_latitude": "3.5952",
        "distance_km": 3,
        "street_address": "Jl. Merdeka No. 1, Medan",
        "street_address_note": "Pagar hijau",
        "label": "Rumah",
        "main": true
      }
    ]
  }
}
```

### `GET` `data.links.next` of `buyer/delivery_addresses`

App method: `AddressApi.page`.

The next page, requested as is when a page came back full. Same response as `GET buyer/delivery_addresses`.

### `POST buyer/delivery_addresses`

App method: `AddressApi.create`.

"Tambah Alamat Baru" (or the first address) at checkout, after the location was detected (activation popup, permission, turn location on, GPS) and the reverse-geocoded address was confirmed or picked on the map. `destination_longitude` / `destination_latitude` come from that location, the other fields from the form; required: `destination_longitude`, `destination_latitude`, `street_address`, `label`. `main: true` makes it the main address. The backend sets the origin from the setting and computes `distance_km`. Answers `201` with the saved address in `data`; a validation error answers `400`, and a point out of the coverage area or already saved answers `422`, with the reason in `message`, shown on the form.

Request:

```http
POST buyer/delivery_addresses
Content-Type: application/json
Authorization: Bearer <JWT>

{
  "destination_longitude": "98.6722",
  "destination_latitude": "3.5952",
  "street_address": "Jl. Merdeka No. 1, Medan",
  "street_address_note": "Pagar hijau",
  "label": "Rumah",
  "main": true
}
```

Response (`AddressResponses.Saved`):

```json
{
  "status_code": 201,
  "status": "Created",
  "message": "",
  "data": {
    "id": "string",
    "record_status": 0,
    "user_id": "string",
    "origin_longitude": "98.6600",
    "origin_latitude": "3.5800",
    "destination_longitude": "98.6722",
    "destination_latitude": "3.5952",
    "distance_km": 3,
    "street_address": "Jl. Merdeka No. 1, Medan",
    "street_address_note": "Pagar hijau",
    "label": "Rumah",
    "main": true
  }
}
```

### `PUT buyer/delivery_addresses/{id}`

App method: `AddressApi.update`.

"Ubah" of an address at checkout: the location is detected again and the form starts from the address's label, note and main flag. Same body and response as `POST buyer/delivery_addresses` (`200`, "delivery address updated successfully"); the backend keeps the saved coordinates and changes only the text fields and `main`. `404` when the address is not the buyer's.

## AuthApi

### `POST buyer/users/sign_in`

App method: `AuthApi.signIn`.

Sign-in. Body: `login` (string) and `password` (string, base64 of the UTF-8 password). The backend matches `login` exactly against `username`, `email` and `mobile_phone` (stored in E.164), so the app sends a phone number typed as `0812…`, `62812…` or `812…` as `+62812…`, and a username or email lower-cased. Answers `201` with `id_token` (the JWT), `expired_at` (RFC 3339 date-time) and `user` (the signed-in user) inside the backend's `data` envelope. Wrong credentials answer `400 login failed`, a locked account `400 user locked`.

Request:

```http
POST buyer/users/sign_in
Content-Type: application/json

{
  "login": "+6281234567890",
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

The app reads the three attributes from `data` (the backend's response envelope) when it is present, and from the top level otherwise. It stores `id_token` as the session JWT, `expired_at` (RFC 3339 date-time) as its expiry, and `user` as the current user. From `user` it reads `id` (the user id) and `mobile_phone_confirmed_at` (null opens the OTP screen, `POST buyer/users/mobile_phone/confirm`); the other `user` fields shown are the ones the backend's user model sends, and are stored as they come. After `expired_at` the app treats the session as signed out and stops sending the token. A 4xx status means the login or password is wrong. In the request, `password` is always the standard base64 (with padding) of the password's UTF-8 bytes, for example `cmFoYXNpYTEyMw==` for `rahasia123`; registration (`POST buyer/users`) encodes its password the same way, while `User/gantiPassword` still sends its passwords as typed.

### `POST buyer/users`

App method: `AuthApi.signUp`.

Buyer registration. Body: `name`, `username`, `mobile_phone`, `password` (base64 of the UTF-8 password) and `referral_code` (all strings). Answers `201` with the new user, shaped like `user` of sign-in, inside the backend's `data` envelope; the app then signs in with the same phone number and password. A 4xx status carries the reason in `message` (e.g. a taken mobile phone or username).

Request:

```http
POST buyer/users
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

The app reads the user from `data` (the backend's response envelope) when it is present, and from the top level otherwise. The backend SMSes a 6-digit confirmation code to `mobile_phone`; the app signs the new buyer in (`POST buyer/users/sign_in`) and opens the OTP screen, which sends the code to `POST buyer/users/mobile_phone/confirm`. A new user's `mobile_phone_confirmed_at` is null. In the request, `password` is the standard base64 of the password's UTF-8 bytes, `referral_code` is left out when empty, and the password confirmation is checked in the app, not sent. A 4xx status means the registration was refused; the app shows the `message` of the error body.

### `POST buyer/users/mobile_phone/confirm`

App method: `AuthApi.confirmMobilePhone`.

Confirms the signed-in buyer's mobile phone with the 6-digit code SMSed at sign-up. Body: `token`, the standard base64 of the code (like passwords). The OTP screen opens after registering (the new buyer is signed in first) and after a sign-in while `user.mobile_phone_confirmed_at` is null; checkout's "verify your number" step uses the same call. Answers `200` with the updated user in `data` (the app stores it), or `200 mobile phone already confirmed` without `data` when there is nothing to confirm. A wrong code answers `422 mobile phone confirmation failed`.

Request:

```http
POST buyer/users/mobile_phone/confirm
Content-Type: application/json
Authorization: Bearer <JWT>

{
  "token": "MTIzNDU2"
}
```

Response:

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "id": "string",
    "role": 2,
    "status": 1,
    "name": "string",
    "username": "string",
    "mobile_phone": "+6281234567890",
    "mobile_phone_confirmed_at": "2026-11-21T16:30:00+07:00"
  }
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

## CartApi

### `GET buyer/cart`

App method: `CartApi.buyerCart`.

The signed-in buyer's cart. The Keranjang screen shows it (signed out: an empty cart without a request), the header badge shows its `quantity`, and every change ("Beli" on the product detail, a new quantity, delete, "Bersihkan Keranjang") reads it before `POST buyer/cart`. Each item carries its product variant's attributes. The app shows `product_photo`, `product_name` (opens the product `product_id`), `quantity` and `stock_unit`, the unit price — `sale_price` when `quantity` ≥ `discount_minimum_quantity` and there is a discount (`selling_price` struck through), otherwise `selling_price` — the line total (unit price × quantity), and a warning when `product_ready_stock` is false; the subtotal is the sum of the line totals. A buyer without a cart gets an empty one (all-zero `id`, no `items`).

Request:

```http
GET buyer/cart
Authorization: Bearer <JWT>
```

Response (`CartResponses.BuyerCartEnvelope`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "id": "string",
    "user_id": "string",
    "quantity": 3,
    "items": [
      {
        "id": "string",
        "product_variant_id": "string",
        "product_category_id": "string",
        "product_category_name": "Sayur",
        "product_id": "string",
        "product_name": "Bayam",
        "product_photo": "https://sadia.id/images/bayam.webp",
        "product_ready_stock": true,
        "stock_unit": "ikat",
        "selling_price": 5000,
        "discount_type": 1,
        "discount_type_name": "persentase",
        "discount_rate": 10,
        "discount_minimum_quantity": 3,
        "discount_amount": 500,
        "sale_price": 4500,
        "description": "string",
        "quantity": 3
      }
    ]
  }
}
```

### `POST buyer/cart`

App method: `CartApi.saveBuyerCart`.

Saves the buyer's whole cart: every listed variant gets the given `quantity` and the cart's other items are deleted. The app always sends the items of `GET buyer/cart` with one change: "Beli" on the product detail (a signed-out buyer logs in first) increases the chosen variant by 1 or appends it with 1; the Keranjang quantity dialog sets a variant's quantity; delete leaves the variant out; "Bersihkan Keranjang" sends `"items": []`. `id` is the cart's id, left out while the cart has none. A variant that is missing or out of stock answers 400 with `items[<i>].product_variant_id: not found or out of stock`. Answers the saved cart, shaped like `GET buyer/cart`; the Keranjang screen shows it and the badge shows its `quantity`.

Request:

```http
POST buyer/cart
Content-Type: application/json
Authorization: Bearer <JWT>

{
  "id": "string",
  "items": [
    {
      "product_variant_id": "string",
      "quantity": 1
    }
  ]
}
```

Response (`CartResponses.BuyerCartEnvelope`): same as `GET buyer/cart`.

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

## CheckoutApi

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

### `GET buyer/promotions/{promo_code}`

App method: `CheckoutApi.buyerPromotion`.

The checkout's "Kode Promo" field. Each edit (trimmed) checks the code, and only the latest answer counts. A code outside `^[a-zA-Z0-9-]+$` is rejected without a request. The code matches exactly (case-sensitive); the path also accepts a promotion id.

- **Valid (200):** the promotion is active, within its dates and under the buyer's `quota_per_buyer`. Its `discount`, a flat amount, comes off the shopping cost (at most what is left after the shopping voucher and category discount). The amount shows as "Potongan Kode Promo", and the order sends `id` as `promotion_id`.
- **Not valid (404 `promotion not found`):** shows "Kode promo tidak berlaku". Any other error shows `message`.

Request:

```http
GET buyer/promotions/{promo_code}
Authorization: Bearer <JWT>
```

Response (`BuyerPromotionResponses.One`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "id": "string",
    "record_status": 1,
    "name": "string",
    "code": "HEMAT-10",
    "deleted_code": null,
    "description": "string",
    "discount": 10000,
    "quota_per_buyer": 2,
    "effective_start_date": "2026-11-01T00:00:00+07:00",
    "effective_end_date": "2026-12-01T00:00:00+07:00",
    "created_by": "string",
    "created_at": "2026-11-01T00:00:00+07:00",
    "updated_by": "string",
    "updated_at": "2026-11-01T00:00:00+07:00",
    "deleted_by": null,
    "deleted_at": null,
    "deletion_reason": null,
    "usages": 1
  }
}
```

## DeliveryApi

### `GET buyer/delivery_plans/now`

App method: `DeliveryApi.deliveryPlanNow`.

The delivery schedule options for a signed-in buyer at checkout: the delivery plan whose transaction window covers now, in the backend's `data` envelope. Each `delivery_times` entry (RFC 3339) is one option: the app shows its day ("Hari ini", "Besok", "Lusa", or e.g. "Sel, 24 Nov"), a session from its hour (Pagi < 11, Siang < 15, Sore < 18, otherwise Malam) and its hours (start to start + `delivery_time_interval`, a Go duration such as `1h0m0s`). A time that has already started cannot be chosen. The chosen option keeps the `delivery_time` text exactly (the transaction is checked against the plan's delivery times by instant) together with `delivery_plan_id`; it expires once that time has passed. No plan for now answers `404 delivery plan not found`: no options. Buyers log in before saving the cart, so there is no signed-out schedule request.

Request:

```http
GET buyer/delivery_plans/now
Authorization: Bearer <JWT>
```

Response (`DeliveryResponses.PlanNow`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "id": "string",
    "record_status": 0,
    "name": "Weekday pagi",
    "weekday": 1,
    "transaction_window_min": "2026-11-20T00:00:00+07:00",
    "transaction_window_max": "2026-11-20T08:00:00+07:00",
    "delivery_time_interval": "1h0m0s",
    "same_day": true,
    "same_day_window_min": "2026-11-20T10:00:00+07:00",
    "same_day_window_max": "2026-11-20T11:00:00+07:00",
    "next_day": true,
    "next_day_window_min": "2026-11-21T08:00:00+07:00",
    "next_day_window_max": "2026-11-21T08:00:00+07:00",
    "next_two_day": false,
    "next_two_day_window_min": null,
    "next_two_day_window_max": null,
    "delivery_times": [
      "2026-11-20T10:00:00+07:00",
      "2026-11-20T11:00:00+07:00",
      "2026-11-21T08:00:00+07:00"
    ]
  },
  "app": "laukpauk"
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

### `GET buyer/payment_channels`

App method: `PaymentApi.buyerPaymentChannels`.

The "Metode Pembayaran" sheet at checkout. `amount` is the total to pay (the checkout grand total); the backend lists only the channels whose minimum and maximum amount allow it, so the app shows every channel it gets. The channels are in `data` (`null` when none). Each row shows `name` (no logo for now); the list is loaded again when the total changes.

Request:

```http
GET buyer/payment_channels?amount=<totalToPay>
Authorization: Bearer <JWT>
```

Response (`PaymentChannelsResponse`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": [
    {
      "id": "string",
      "name": "BCA Virtual Account",
      "channel_code": "BCA_VIRTUAL_ACCOUNT"
    }
  ]
}
```

### `POST buyer/transactions`

App method: `PaymentApi.createTransaction`.

"Buat Pesanan" in the "Metode Pembayaran" sheet, after a channel is picked. The body (`PaymentRepository.transactionBody`): `cart_id` and `items[]` (`cart_item_id`, `buyer_note`) of `GET buyer/cart`; `recipient_name`, `recipient_mobile_phone` and `buyer_note` from the checkout form; `delivery_address_id` of the chosen delivery address; `delivery_datetime`, the chosen slot's `delivery_time` exactly as `GET buyer/delivery_plans/now` sent it; `reward_points_discounted`, or `voucher_id` when a voucher is chosen; `promotion_id` when a valid promo code is entered; `wallet_discounted` and `wallet_amount` (`wallet` of `GET buyer/users/me`); `payment_channel_id`, the `id` of the picked channel. On `201` the order success screen shows `reference_number`, `payment_amount`, `payment_channel_name`, `payment_virtual_account_number` and `payment_due_date` ("Pesanan sudah lunas." when `payment_amount` is 0); an error shows `message` in a dialog.

Request:

```http
POST buyer/transactions
Content-Type: application/json
Authorization: Bearer <JWT>

{
  "cart_id": "string",
  "items": [
    {
      "cart_item_id": "string",
      "buyer_note": "Potong kecil"
    }
  ],
  "recipient_name": "Budi",
  "recipient_mobile_phone": "081234567890",
  "buyer_note": "Titip satpam",
  "delivery_address_id": "string",
  "delivery_datetime": "2026-11-26T07:00:00+07:00",
  "reward_points_discounted": false,
  "voucher_id": "string",
  "promotion_id": "string",
  "wallet_discounted": true,
  "wallet_amount": 10000,
  "payment_channel_id": "string"
}
```

Response (`TransactionResponses.Created`), the fields the app reads:

```json
{
  "status_code": 201,
  "status": "Created",
  "message": "transaction created successfully",
  "data": {
    "id": "string",
    "reference_number": "26330-1-ABCD",
    "status": 0,
    "payment_amount": 54321,
    "payment_due_date": "2026-11-26T00:34:00+07:00",
    "payment_channel_id": "string",
    "payment_channel_name": "BCA Virtual Account",
    "payment_virtual_account_number": "1234567890",
    "payment_virtual_account_expired_at": null
  }
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

### `GET buyer/products/{id}`

App method: `ProductApi.buyerProduct`.

The product detail screen, for every `product_type` (paket included). The product is in the backend's `data` envelope, with its active `variants`. The app shows `photo` (full URL), `name`, `category_name`, `description`, `ready_stock` (false hides "Beli") and per variant `stock_unit`, `sale_price`, `selling_price` (struck through when discounted), `discount_type` (0 = fixed amount "cash", 1 = percentage "persentase"), `discount_rate`, `discount_amount`, `discount_minimum_quantity` and `description`. A missing or inactive product answers 404. The other fields are ignored.

Request:

```http
GET buyer/products/01a1f1aa-c5ff-73dc-9618-01b224bed65f
```

Response (`ProductResponses.BuyerProductDetail`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
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
    "hits": 1,
    "lowest_selling_price": 37000,
    "variants": [
      {
        "id": "string",
        "record_status": 0,
        "product_category_id": "string",
        "product_category_name": "string",
        "product_id": "string",
        "product_name": "string",
        "product_photo": "string",
        "product_ready_stock": true,
        "stock_unit": "string",
        "selling_price": 37000,
        "discount_type": 0,
        "discount_type_name": "cash",
        "discount_rate": 0,
        "discount_minimum_quantity": 0,
        "discount_amount": 0,
        "sale_price": 37000,
        "description": "string"
      }
    ]
  },
  "app": "laukpauk"
}
```

### `GET Produk/detailProduk`

App method: `ProductApi.productDetail`.

Now only used by the "Beli" unit sheet (`ProductRepository.loadUnitSheetData`); the detail screen uses `GET buyer/products/{id}`.

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

### `GET buyer/recipes`

App method: `RecipeApi.buyerRecipes`.

The Resep page: the active recipes, newest first, 25 per page (`limit=25`). The search box sends the name as `q` (case-insensitive match). Page-based: `data.links.next` (empty on the last page) is loaded when the list is scrolled to the end; the recipes are `data.rows`. Each card shows `name`, `photo` (a full URL, used as is), `description` (cut at 30 characters) and the `instagram_link` / `facebook_link` buttons.

Request:

```http
GET buyer/recipes?limit=25&q=<q>
```

Response (`BuyerRecipeResponses.Page`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "links": {
      "first": "https://sadia.id/v1/buyer/recipes?limit=25&page=1",
      "previous": "",
      "current": "https://sadia.id/v1/buyer/recipes?limit=25&page=1",
      "next": "",
      "last": "https://sadia.id/v1/buyer/recipes?limit=25&page=1"
    },
    "info": {
      "limit": 25,
      "page": 1,
      "pages": 1,
      "total": 1
    },
    "rows": [
      {
        "id": "string",
        "record_status": 1,
        "photo": "https://sadia.id/photos/r1.jpg",
        "name": "Sayur Asem",
        "deleted_name": null,
        "description": "string",
        "instagram_link": "https://instagram.com/p/...",
        "facebook_link": "",
        "created_by": "string",
        "created_at": "2026-11-25T21:00:00+07:00",
        "updated_by": "string",
        "updated_at": "2026-11-25T21:00:00+07:00",
        "deleted_by": null,
        "deleted_at": null,
        "deletion_reason": null
      }
    ]
  }
}
```

### `GET` `data.links.next` of `buyer/recipes`

App method: `RecipeApi.buyerRecipesPage`.

The next page, requested as is when the list is scrolled to the end. Same response as `GET buyer/recipes`.

### `GET buyer/recipes/{id}`

App method: `RecipeApi.buyerRecipe`.

The recipe detail: one active recipe with its active `guides` and `items`, both in ascending order. 404 when the recipe isn't there, which shows the "not found" view. The screen shows:

- `photo` (a full URL, used as is), `name` and `description`, plus the `instagram_link` / `facebook_link` buttons.
- The `guides` as numbered "Cara Memasak" steps.
- Each item as an ingredient row: `product_name`, then "Diperlukan `quantity` `product_variant_stock_unit`" with `note` in brackets. When `product_ready_stock` is false the row shows "Stok habis".
- A variant picker for each in-stock item, loaded from `GET buyer/products/{id}` (`ProductApi.buyerProduct`, `data.variants`). It starts on `product_variant_id` and `quantity`.

"Belanja Resep" adds the picked variants and quantities to the cart with one read-then-save on `GET buyer/cart` / `POST buyer/cart` (`CartApi`).

Request:

```http
GET buyer/recipes/{id}
```

Response (`BuyerRecipeResponses.One`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "id": "string",
    "record_status": 1,
    "photo": "https://sadia.id/photos/r1.jpg",
    "name": "Sayur Asem",
    "description": "string",
    "instagram_link": "https://instagram.com/p/...",
    "facebook_link": "",
    "guides": [
      {
        "id": "string",
        "record_status": 1,
        "recipe_id": "string",
        "instruction": "Rebus air hingga mendidih"
      }
    ],
    "items": [
      {
        "id": "string",
        "record_status": 1,
        "recipe_id": "string",
        "product_id": "string",
        "product_name": "Asam Jawa",
        "product_photo": "https://sadia.id/photos/p1.jpg",
        "product_ready_stock": true,
        "product_variant_id": "string",
        "product_variant_stock_unit": "gram",
        "quantity": 2,
        "note": "secukupnya"
      }
    ]
  }
}
```

## SettingApi

### `GET buyer/setting`

App method: `SettingApi.buyerSetting`.

The store settings (no sign-in needed). Checkout's delivery cost is `distance_km` of the delivery address (`GET buyer/delivery_addresses`, the address at the saved delivery point) times `delivery_cost_per_km`; the product unit sheet shows `delivery_cost_per_km` as the per-km rate. The app reads `delivery_cost_per_km` and `delivery_max_distance_km`.

Request:

```http
GET buyer/setting
```

Response (`SettingResponses.Buyer`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "vacation_mode": false,
    "vacation_end_date": null,
    "voucher_enabled": true,
    "referral_points_enabled": false,
    "referral_points_minimum_shopping_cost": 0,
    "referral_points_amount": 0,
    "referral_points_slots": [],
    "referral_points_expense": 0,
    "reward_points_accumulation_window_in_days": 0,
    "delivery_origin_longitude": "98.6600",
    "delivery_origin_latitude": "3.5800",
    "delivery_cost_per_km": 5000,
    "delivery_max_distance_km": 20,
    "whatsapp_phone_number": "+6281234567890",
    "payment_terms": "1h0m0s",
    "maximum_delivery_time": "3h0m0s",
    "transaction_auto_withdrawal_window_min": null,
    "transaction_auto_withdrawal_window_max": null,
    "otp_required_for_login": false
  }
}
```

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

### `GET buyer/transactions`

App method: `TransactionApi.buyerTransactions`.

"Pesanan Saya": the buyer's transactions, newest first, 25 per page. Query: `limit=25`; on "Hari Ini" (the first tab), `created_at_start` / `created_at_end` cover today and `status` comes from its dropdown (none for "Semua Status", else 0..8); on the other tabs `status` comes from the tab (Semua: none, Belum Bayar 0, Dibayar 2, Diproses 5, Dikirim 6, Selesai 8, Dibatalkan 1) and `created_at_start` / `created_at_end` from the date range, 00:00:00 of the start day and 23:59:59 of the end day in the phone's time zone, sent as RFC 3339 UTC (`…Z`); either may be left out. Page-based: `data.links.next` (empty on the last page) is loaded when the list is scrolled to the end; the transactions are `data.rows`. Each card shows `reference_number`, `payment_channel_name`, the delivery day and time of `delivery_datetime`, `created_at`, the status and `revised_payment_amount`.

Request:

```http
GET buyer/transactions?limit=25&status=<status>&created_at_start=<RFC3339>&created_at_end=<RFC3339>
Authorization: Bearer <JWT>
```

Response (`BuyerTransactionResponses.Page`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "links": {
      "first": "https://sadia.id/v1/buyer/transactions?limit=25&page=1",
      "previous": "",
      "current": "https://sadia.id/v1/buyer/transactions?limit=25&page=1",
      "next": "https://sadia.id/v1/buyer/transactions?limit=25&page=2",
      "last": "https://sadia.id/v1/buyer/transactions?limit=25&page=2"
    },
    "info": {
      "limit": 25,
      "page": 1,
      "pages": 2,
      "total": 30
    },
    "rows": [
      {
        "id": "string",
        "reference_number": "LP26330001ABCD",
        "status": 6,
        "buyer_note": "Titip satpam",
        "unique_code": 123,
        "shopping_cost": 50123,
        "voucher_type": 0,
        "voucher_discount": 5000,
        "reward_points_discount": 0,
        "wallet_discount": 10000,
        "delivery_estimated_cost": 15000,
        "delivery_street_address": "Jl. Merdeka No. 1, Medan",
        "delivery_street_address_note": "Pagar hijau",
        "delivery_datetime": "2026-11-26T07:00:00+07:00",
        "payment_amount": 50123,
        "payment_due_date": null,
        "payment_channel_name": "BCA Virtual Account",
        "payment_virtual_account_number": null,
        "created_at": "2026-11-25T21:00:00+07:00",
        "withdrawal_reason": null,
        "revised_payment_amount": 45123,
        "refund_amount": 5000,
        "items": [
          {
            "id": "string",
            "product_name": "Bayam",
            "stock_unit": "ikat",
            "selling_price": 5000,
            "quantity": 2,
            "subtotal": 10000,
            "buyer_note": "Segar",
            "revised_quantity": 1,
            "revised_subtotal": 5000,
            "revision_note": "Stok tinggal 1",
            "out_of_stock": false
          }
        ]
      }
    ]
  }
}
```

### `GET` `data.links.next` of `buyer/transactions`

App method: `TransactionApi.buyerTransactionsPage`.

The next page, requested as is when the list is scrolled to the end. Same response as `GET buyer/transactions`.

### `GET buyer/transactions/{id}`

App method: `TransactionApi.buyerTransaction`.

"Detail Transaksi" and its "Daftar Belanjaan" (also opened from the order success screen). It shows the status, `reference_number`, `payment_channel_name`, `buyer_note`, `delivery_street_address` (+ `delivery_street_address_note`), the delivery day and time, and the breakdown: items total (`shopping_cost` − `unique_code`), the voucher on the shopping or delivery cost (`voucher_type`), points, wallet, delivery cost, unique code, `payment_amount`, and when different the revised total and `refund_amount`; while unpaid also `payment_virtual_account_number` and `payment_due_date`. Cancelled or withdrawn orders show `withdrawal_reason`. Each item shows its name, quantity and unit, `revised_subtotal` (struck through when `out_of_stock`), the revised quantity when it changed, and the buyer's and seller's notes. `404` shows "not found".

Request:

```http
GET buyer/transactions/{id}
Authorization: Bearer <JWT>
```

Response (`BuyerTransactionResponses.One`): one transaction, the same attributes as a row of `GET buyer/transactions`, in `data`.

### `GET Transaksi/getTransaksi/{id}`

App method: `TransactionApi.getTransaction`.

Still used by "Belanja Lagi" (repeat order) for the transaction number and delivery date.

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

## UserApi

### `GET buyer/users/me`

App method: `UserApi.me`.

The signed-in buyer, used by the Akun dashboard, Profil Saya, Kode Referal, the home wallet card (wallet and points), checkout (name, phone, address, wallet and points in one request) and the auth screens (phone number). The app reads `name`, `username`, `address`, `mobile_phone`, `mobile_phone_confirmed_at`, `photo` (a full URL, or a legacy file name in the user folder), `referral_code`, `created_at` (shown as "24 November 2026"), `wallet` (the dompet balance) and `reward_points` (the buyer's points); null is shown as 0. Replaces `User/getDetailUser`, `User/getTotalDompet` and `Point/getPoint`.

Request:

```http
GET buyer/users/me
Authorization: Bearer <JWT>
```

Response (`AccountResponses.Me`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "id": "string",
    "role": "buyer",
    "name": "Budi",
    "mobile_phone_confirmed_at": "2026-11-21T16:30:00+07:00",
    "mobile_phone": "+6281234567890",
    "username": "budi",
    "email": null,
    "password_changed_at": null,
    "sign_in_count": 3,
    "current_sign_in_at": "2026-11-24T09:00:00+07:00",
    "last_sign_in_at": "2026-11-23T20:00:00+07:00",
    "current_sign_in_ip": "string",
    "last_sign_in_ip": "string",
    "failed_attempts": 0,
    "locked_at": null,
    "photo": "",
    "user_level": "string",
    "referral_code": "BUDI01",
    "address": "Jl. Merdeka No. 1",
    "referrer_id": null,
    "referral_points": 0,
    "reward_points": 10,
    "wallet": 25000,
    "created_at": "2026-11-19T10:00:00+07:00"
  },
  "app": "laukpauk"
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

### `GET buyer/vouchers`

App method: `VoucherApi.buyerVouchers`.

Used in three places:

- "Voucher Saya" (from the home page): without `purchase`, all the buyer's active vouchers.
- The voucher count on the home page wallet card and the account dashboard: the length of `data` with `purchase` = the cart grand total (the total of `GET buyer/cart`); `404` counts as 0.
- The "Voucher Kamu" sheet at checkout, described below.

At checkout, `purchase` is the cart grand total (the shopping total of `GET buyer/cart`); the backend lists the buyer's active vouchers whose `minimum_purchase` is at most `purchase`. The vouchers are in `data` (`404` shows the empty state). Each row shows `name`, `photo`, the type ("ongkir" for `voucher_type` 0, "belanja" for 1), `discount` and `minimum_purchase`. The app checks only that the shopping cost is at least `minimum_purchase`; otherwise choosing the voucher shows "Voucher Tidak Dapat Digunakan" with the minimum. Dates and times are left to the backend. `voucher_type` 0 reduces the delivery cost, any other type the shopping cost, by `discount` up to that cost. Points and a voucher exclude each other, and the picker shows only when `voucher_enabled` of `GET buyer/setting` is true.

Request:

```http
GET buyer/vouchers?purchase=<cartGrandTotal>
Authorization: Bearer <JWT>
```

Response (`CheckoutResponses.VoucherList`):

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": [
    {
      "id": "string",
      "record_status": 1,
      "voucher_type": 0,
      "voucher_type_name": "delivery cost",
      "target_audience": 0,
      "target_audience_name": "all buyers",
      "code": "ONGKIRHEMAT",
      "deleted_code": null,
      "photo": "https://sadia.id/...",
      "name": "Gratis Ongkir",
      "discount": 10000,
      "minimum_purchase": 50000,
      "quota_per_buyer": 1,
      "effective_start_time": "0001-01-01T00:00:00Z",
      "effective_end_time": "0001-01-01T23:59:59Z",
      "effective_start_date": "2026-11-01T00:00:00Z",
      "effective_end_date": "2026-11-30T00:00:00Z",
      "description": "string",
      "created_by": "string",
      "created_at": "2026-11-01T00:00:00Z",
      "updated_by": "string",
      "updated_at": "2026-11-01T00:00:00Z",
      "deleted_by": null,
      "deleted_at": null,
      "deletion_reason": null
    }
  ]
}
```

### `GET buyer/vouchers/{id}`

App method: `VoucherApi.buyerVoucher`.

"Detail Voucher" (opened from "Voucher Saya" or a push notification): one of the buyer's active vouchers. It shows `name`, `photo`, the type, `discount`, `code`, `minimum_purchase`, `quota_per_buyer` as the maximum use, the effective dates and times with their zone (e.g. "pukul 08:00 - 17:00 WIB" for +07:00), and `description`. `404` shows "Detail Voucher Tidak Ditemukan".

Request:

```http
GET buyer/vouchers/{id}
Authorization: Bearer <JWT>
```

Response (`VoucherResponses.BuyerVoucher`): one voucher, the same attributes as a row of `GET buyer/vouchers`, in `data`.

```json
{
  "status_code": 200,
  "status": "OK",
  "message": "",
  "data": {
    "id": "string",
    "voucher_type": 0,
    "voucher_type_name": "delivery cost",
    "code": "ONGKIRHEMAT",
    "photo": "string",
    "name": "Gratis Ongkir",
    "discount": 10000,
    "minimum_purchase": 50000,
    "quota_per_buyer": 1,
    "effective_start_time": "0001-01-01T08:00:00+07:00",
    "effective_end_time": "0001-01-01T17:00:00+07:00",
    "effective_start_date": "2026-11-01T00:00:00+07:00",
    "effective_end_date": "2026-11-30T00:00:00+07:00",
    "description": "string"
  }
}
```
