# ==============================================================================
# 用来进行代码生成的 Makefile
#

gen.add-copyright: tools.verify.addlicense ## 添加版权头信息.
	@addlicense -v -f $(PROJ_ROOT_DIR)/scripts/boilerplate.txt $(PROJ_ROOT_DIR) --skip-dirs=third_party,vendor,$(OUTPUT_DIR)

gen.ca: ## 生成 CA 文件.
	@mkdir -p $(OUTPUT_DIR)/cert
	@# 1. 生成根证书私钥 (CA Key)
	@openssl genrsa -out $(OUTPUT_DIR)/cert/ca.key 4096
	@# 2. 使用根私钥生成证书签名请求 (CA CSR)，有效期为 10 年
	@openssl req -new -nodes -key $(OUTPUT_DIR)/cert/ca.key -sha256 -days 3650 -out $(OUTPUT_DIR)/cert/ca.csr \
		-subj "/C=CN/ST=Guangdong/L=Shenzhen/O=onexstack/OU=it/CN=127.0.0.1/emailAddress=stary99c@163.com"
	@# 3. 使用根私钥签发根证书 (CA CRT)，使其自签名
	@openssl x509 -req -days 365 -in $(OUTPUT_DIR)/cert/ca.csr -signkey $(OUTPUT_DIR)/cert/ca.key -out $(OUTPUT_DIR)/cert/ca.crt
	@# 4. 生成服务端私钥
	@openssl genrsa -out $(OUTPUT_DIR)/cert/server.key 2048
	@# 5. 使用服务端私钥生成服务端的证书签名请求 (Server CSR)
	@openssl req -new -key $(OUTPUT_DIR)/cert/server.key -out $(OUTPUT_DIR)/cert/server.csr \
		-subj "/C=CN/ST=Guangdong/L=Shenzhen/O=serverdevops/OU=serverit/CN=localhost/emailAddress=stary99c@163.com" \
		-addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
	@# 6. 使用根证书 (CA) 签发服务端证书 (Server CRT)
	@openssl x509 -days 365 -sha256 -req -CA $(OUTPUT_DIR)/cert/ca.crt -CAkey $(OUTPUT_DIR)/cert/ca.key \
		-CAcreateserial -in $(OUTPUT_DIR)/cert/server.csr -out $(OUTPUT_DIR)/cert/server.crt -extensions v3_req \
		-extfile <(printf "[v3_req]\nsubjectAltName=DNS:localhost,IP:127.0.0.1")

gen.protoc: tools.verify.buf ## 编译 protobuf 文件（使用 buf）.
	@echo "===========> Generate protobuf files with buf"
	@mkdir -p $(PROJ_ROOT_DIR)/api/openapi
	@buf generate --exclude-path third_party --exclude-path pkg/api/google --exclude-path pkg/api/protoc-gen-openapiv2 --exclude-path pkg/api/github.com
	@find $(APIROOT)/apiserver -name "*.pb.go" -exec protoc-go-inject-tag -input={} \;

gen.generate:
	@GOWORK=off go generate ./...

# 伪目标（防止文件与目标名称冲突）
.PHONY: gen.add-copyright gen.ca gen.protoc gen.generate
