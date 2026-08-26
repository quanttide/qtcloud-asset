import 'login_redirect_io.dart'
    if (dart.library.js_interop) 'login_redirect_web.dart';

typedef LoginRedirect = void Function(Uri uri);

void defaultLoginRedirect(Uri uri) => defaultLoginRedirectImpl(uri);
