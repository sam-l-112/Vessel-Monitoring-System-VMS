# nginx 
## nginx config setting 
file 位置 :
```bash
/etc/nginx/nginx.conf
```
or 
```bash
/etc/nginx/sites-available/
```

## 不同域名 config file
```bash
/etc/nginx/conf.d/*.conf
```
```bash
/etc/nginx/conf.d/kylemocode.com.conf
```
## port 網頁設定
```bash
sudo nano /etc/nginx/sites-available/default
```

## 1. 檢查語法是否有誤
```bash
sudo nginx -t
```

## 2. 重新載入設定
```bash
sudo systemctl reload nginx
```

## 啟動
```bash
sudo systemctl start nginx
```

## 查看狀態
```bash
sudo systemctl status nginx
```
## port nginx 檢查
```bash
sudo ss -tunlp | grep -E ':3000|:8080'
```

---
# download
## Install the prerequisites:

```bash
sudo apt install curl gnupg2 ca-certificates lsb-release ubuntu-keyring
```

## Import an official nginx signing key so apt could verify the packages authenticity. Fetch the key:

```bash
curl https://nginx.org/keys/nginx_signing.key | gpg --dearmor \
    | sudo tee /usr/share/keyrings/nginx-archive-keyring.gpg >/dev/null
```
## Verify that the downloaded file contains the proper key:

```bash
gpg --dry-run --quiet --no-keyring --import --import-options import-show /usr/share/keyrings/nginx-archive-keyring.gpg
```

## The output should contain the full fingerprint 573BFD6B3D8FBC641079A6ABABF5BD827BD9BF62 as follows:

```bash
pub   rsa2048 2011-08-19 [SC] [expires: 2027-05-24]
      573BFD6B3D8FBC641079A6ABABF5BD827BD9BF62
uid                      nginx signing key <signing-key@nginx.com>
```

## Note that the output can contain other keys used to sign the packages.

## Set up repository pinning to prefer our packages over distribution-provided ones:
```bash
echo -e "Package: *\nPin: origin nginx.org\nPin: release o=nginx\nPin-Priority: 900\n" \
    | sudo tee /etc/apt/preferences.d/99nginx
```
## To install nginx, run the following commands:
```bash
sudo apt update
sudo apt install nginx
```