# mariadb 
# 



---
# download

## Update Package List:

Before installing, it's a good practice to update your package index.

For Debian/Ubuntu:Bash

```bash
sudo apt update
```

## Install MariaDB Server:

Install the MariaDB server and client packages.

For Debian/Ubuntu:Bash

```bash
sudo apt install mariadb-server mariadb-client galera-4
```


## Secure the Installation:

After installation, run the security script to set a root password, remove anonymous users, and disable remote root login.


```bash
sudo mariadb-secure-installation
```
Follow the prompts to configure your security settings.

## Start and Verify the Service:

MariaDB typically starts automatically after installation. You can check its status and manually start it if needed.

Check status:


```bash
sudo systemctl status mariadb
```
## Start service (if not running):Bash


```bash
sudo systemctl start mariadb
```
Verify installation by connecting as root:Bash


```bash
mariadb -u root -p
```
Enter the root password you set during the secure installation.