#!/bin/bash
set -ex

# 生成 CA 的私钥
openssl genrsa -aes256 -passout pass:1 -out ca.key.pem 4096
openssl rsa -passin pass:1 -in ca.key.pem -out ca.key.pem.tmp
mv ca.key.pem.tmp ca.key.pem

# 生成自签名证书并输出为 ca.crt
openssl req -config openssl.cnf -key ca.key.pem -new -x509 -days 7300 -sha256 -extensions v3_ca -out ca.crt
