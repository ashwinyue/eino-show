#!/bin/bash
curl https://127.0.0.1:5555/healthz
# 3. 通过 HTTPS 协议访问 /healthz，指定根证书
curl https://127.0.0.1:5555/healthz --ciphers DEFAULT@SECLEVEL=1 --cacert $HOME/.miniblog/cert/ca.crt
# 4. 忽略 HTTPS 证书参数，指定跳过 SSL 检测
curl https://127.0.0.1:5555/healthz -k
