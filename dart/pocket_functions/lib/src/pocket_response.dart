// Main library, extract later
import 'dart:convert';

class PocketResponse {
  final buffer = StringBuffer();
  final headers = <String, String>{};

  PocketResponse addHeader(String key, String value) {
    headers[key] = value;
    return this;
  }

  PocketResponse write(Object s) {
    if (_isJsonResponse()) {
      buffer.write(jsonEncode(s));
    } else {
      buffer.write(s);
    }
    return this;
  }

  PocketResponse writeln(String s) {
    buffer.writeln(s);
    return this;
  }

  void close() {}

  @override
  String toString() {
    final out = StringBuffer();
    headers.forEach((key, value) {
      out.writeln("$key=$value");
    });
    out.writeln("====");
    out.write(buffer.toString());
    return out.toString();
  }

  bool _isJsonResponse() {
    return headers.containsKey("content-type") &&
        headers["content-type"] == "application/json";
  }
}
