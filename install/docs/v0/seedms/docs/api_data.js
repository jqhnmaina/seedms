define({ "api": [
  {
    "type": "get",
    "url": "/docs",
    "title": "Docs",
    "name": "Docs",
    "version": "0.1.0",
    "group": "Service",
    "success": {
      "fields": {
        "200": [
          {
            "group": "200",
            "type": "html",
            "optional": false,
            "field": "docs",
            "description": "<p>Docs page to be viewed on browser.</p>"
          }
        ]
      }
    },
    "filename": "pkg/handler/http/handler.go",
    "groupTitle": "Service"
  },
  {
    "type": "get",
    "url": "/status",
    "title": "Status",
    "name": "Status",
    "version": "0.1.0",
    "group": "Service",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "200": [
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>Micro-service name.</p>"
          },
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "version",
            "description": "<p>http://semver.org version.</p>"
          },
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "description",
            "description": "<p>Short description of the micro-service.</p>"
          },
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "canonicalName",
            "description": "<p>Canonical name of the micro-service.</p>"
          }
        ]
      }
    },
    "filename": "pkg/handler/http/handler.go",
    "groupTitle": "Service"
  }
] });
