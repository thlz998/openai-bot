# 企业微信版本chatgpt机器人

## 快速运行

``` bash
docker run -d -p 3201:3200 \
  -e api_key="[chatgpt的APIkey]" \
  -e app_port=":3200" \
  -e app_model="text-davinci-003" \
  -e wework_token="[企业微信自建应用的API token]" \
  -e wework_encodingAeskey="[企业微信自建应用的encodingAeskey]" \
  -e wework_corpid="[企业微信企业唯一ID]" \
  -e wework_secret="[企业微信自建应用的secret]" \
  -e wework_agentid="[企业微信自建应用的agentid]" \
  ibfpig/openai-wework:1.0.0
```

## 自行构建docker image

``` bash
docker build -t 镜像名称:tag .
```

## 使用docker-compose.yml运行

1. 修改`docker-compose.yml`文件中的环境变量

2. 运行

``` bash
docker-compose up -d
```

## TODO

- [x] 支持企业微信openai
- [ ] 支持官网版本的ChatGPT
- [ ] 优化企业微信token获取方式(目前比较粗暴，没存，每次使用会重新获取)
- [ ] 兼容飞书、钉钉等其他企业工具
