import 'dart:js_interop';

@JS('window.location.assign')
external void _assign(JSString href);

void defaultLoginRedirectImpl(Uri uri) {
  _assign(uri.toString().toJS);
}
