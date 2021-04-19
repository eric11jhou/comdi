# comdi

```bash
make build_linux_amd64
docker build -t comdi:latest --no-cache -f Dockerfile .
docker run -i -t -p 80:80 comdi:latest /bin/sh
docker tag comdi:latest eric11jhou/comdi:latest
docker push eric11jhou/comdi:latest
```