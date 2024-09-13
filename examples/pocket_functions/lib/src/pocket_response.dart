import 'dart:convert';
import 'pocket_request.dart';

class PocketRequest {
  final String path;
  final String httpMethod;
  final String? body;
  final String contentType;
  final Map<String, List<String>> params;

  final response = PocketResponse();

  PocketRequest({
    required this.path,
    required this.httpMethod,
    required this.body,
    required this.contentType,
    required this.params,
  });

  factory PocketRequest.fromEnvironment(Map<String, String> env) {
    // From dart.go
    // env["pf_path"] = path
    // env["pf_query"] = c.Request.URL.RawQuery
    // env["pf_method"] = c.Request.Method
    // 	env["pf_body"] = string(body)
    // env["pf_content-type"] = c.GetHeader("Content-Type")
    Uri uri = Uri(query: env["pf_query"]);

    return PocketRequest(
      path: env["pf_path"]!,
      httpMethod: env["pf_method"]!,
      body: env['pf_body'],
      contentType: env['pf_content-type'] ?? "text/plain",
      params: uri.queryParametersAll,
    );
  }

  String? plainBody() {
    return body;
  }

  Map<String, dynamic>? jsonBody() {
    if (body == null) {
      return null;
    }

    if (contentType != "application/json") {
      return null;
    } else {
      return jsonDecode(body!);
    }
  }
}
