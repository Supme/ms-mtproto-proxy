# ms-mtproto-proxy
Multi secrets code and realtime refresh for https://github.com/TelegramMessenger/MTProxy

Download from https://github.com/Supme/ms-mtproto-proxy/releases

create /opt/MTProto/secrets.conf example
```
a1ef210a6a2b07284b148b3dc532c9be username@example.tld
#3f110568973d3189e7506d6dc73a4aa0 commenteduser@notactive.tld
```

Execute:
```
ExecStart=/opt/MTProxy/ms-mtproto-proxy ./secrets.conf ./mtproto-proxy -u nobody -p 8888 ...other you keys... --aes-pwd proxy-secret proxy-multi.conf -M 1
```

Stop you MTProxy and change you /etc/systemd/system/MTProxy.service
```
[Unit]
Description=MTProxy
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/MTProxy
ExecStart=/opt/MTProxy/ms-mtproto-proxy ./secrets.conf ./mtproto-proxy -u nobody -p 8888 ...other you keys... --aes-pwd proxy-secret proxy-multi.conf -M 1
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

```
systemctl daemon-reload
```
and start you MTProxy

later need only edit secrets.conf and not need restart service
