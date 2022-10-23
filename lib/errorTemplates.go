package app_lib

// TODO временное решение, удобно только для GUI

const errorTemplate500 = `<!DOCTYPE html>
<html lang="en">
  <head>
    <!-- Title -->
    <title>500</title>

    <!-- Required Meta Tags Always Come First -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <meta http-equiv="x-ua-compatible" content="ie=edge">

    <!-- Favicon -->
    <link rel="shortcut icon" href="favicon.ico">

    <!-- Google Fonts -->
    <link rel="stylesheet" href="//fonts.googleapis.com/css?family=Open+Sans%3A400%2C300%2C500%2C600%2C700">

    <!-- CSS Global Compulsory -->
    <link rel="stylesheet" href="/templates/unify-v2.5.1/html/assets/vendor/bootstrap/bootstrap.min.css">

    <!-- CSS Unify -->
    <link rel="stylesheet" href="/templates/unify-v2.5.1/html/assets/css/unify-core.css">
    <link rel="stylesheet" href="/templates/unify-v2.5.1/html/assets/css/unify-components.css">
    <link rel="stylesheet" href="/templates/unify-v2.5.1/html/assets/css/unify-globals.css">

    <!-- CSS Customization -->
    <link rel="stylesheet" href="assets/css/custom.css">
  </head>

  <body>
    <main>
      <div class="container g-py-150 text-center">
        <h1 class="g-font-size-40 mb-4">Ошибка сервера (500)</h1>
        <p class="g-font-size-16 mb-5">Мы уже знаем об этой ошибке и уже исправляем её. Попробуйте чуть позже.</p>
        <p class="g-font-size-14 mb-5">{{.}}</p>
        <a class="btn u-btn-outline-primary g-font-size-13 text-uppercase g-py-10 g-px-15"  onclick="history.back();return false;" href="#">Вернуться</a>
      </div>
    </main>

    <!-- JS Global Compulsory -->
    <script src="/templates/unify-v2.5.1/html/assets/vendor/jquery/jquery.min.js"></script>
    <script src="/templates/unify-v2.5.1/html/assets/vendor/jquery-migrate/jquery-migrate.min.js"></script>
    <script src="/templates/unify-v2.5.1/html/assets/vendor/popper.min.js"></script>
    <script src="/templates/unify-v2.5.1/html/assets/vendor/bootstrap/bootstrap.min.js"></script>

    <!-- JS Unify -->
    <script src="/templates/unify-v2.5.1/html/assets/js/hs.core.js"></script>

    <!-- JS Custom -->
    <script src="/templates/unify-v2.5.1/html/assets/js/custom.js"></script>
  </body>
</html>
`

const errorTemplate503 = `<div class="container text-center">
        <h3 class="h5 u-heading-v3__title g-font-primary g-font-weight-700 g-color-gray-dark-v1 text-uppercase g-brd-primary">Ошибка (503)</h3>
        <p class="g-font-size-14 g-py-10">{{.}}</p>
      </div>
`
