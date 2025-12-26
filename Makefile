swaggerUI:
	rm -rf api/swagger
	curl -L https://github.com/swagger-api/swagger-ui/archive/refs/tags/v5.31.0.tar.gz -o /tmp/swagger-api.tar.gz
	cd /tmp && tar xzvf swagger-api.tar.gz
	cp -r /tmp/swagger-ui-5.31.0/dist api/swagger
	sed -i api/swagger/swagger-initializer.js -e 's|url: "https://petstore.swagger.io/v2/swagger.json",|url: "openapi.yaml",|'
