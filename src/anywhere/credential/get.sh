# 生成CA根证书
openssl genrsa -out ca.key 2048
# 使用-x509参数生成证书
openssl req -new -x509 -days 7200 -key ca.key -out ca.pem

# 生成服务端证书
# openssl genrsa -out server.key 2048
openssl ecparam -genkey -name secp384r1 -out server.key
# 生成证书请求文件
openssl req -new -key server.key -out server.csr
# 基于CA签发服务端证书
openssl x509 -req -sha256 -CA ca.pem -CAkey ca.key -CAcreateserial -days 3650 -in server.csr -out server.pem


# 生成客户端证书
openssl ecparam -genkey -name secp384r1 -out client.key
openssl req -new -key client.key -out client.csr
openssl x509 -req -sha256 -CA ca.pem -CAkey ca.key -CAcreateserial -days 3650 -in client.csr -out client.pem

