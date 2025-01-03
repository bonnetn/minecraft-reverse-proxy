# Minecraft Reverse Proxy

A simple reverse proxy for Minecraft servers that allows multiple Minecraft server instances to run on the same IP address but on different hostnames. Despite being only **~250 lines of code**, this proxy efficiently listens for incoming connections and routes them to the correct backend server based on the hostname provided by the client.


## Features

- **Single IP, Multiple Servers**: No need to purchase or use multiple IP addresses. The proxy routes traffic based on hostname.
- **Easy Configuration**: All configuration is done via environment variables.
- **Lightweight**: Minimal overhead, designed to be fast and efficient.

## How It Works

When a Minecraft client connects to your server (e.g., `some.hostname.com`), the proxy intercepts the connection on a single IP and port (for example, `:25565`). It then looks up the corresponding backend server based on the hostname, and routes the connection there.

## Usage

1. **Set the `PROXY_MAPPING` environment variable** to define your hostname-to-server mappings:

   ```bash
   PROXY_MAPPING={"default": "localhost:25565", "servers": {"some.hostname.com": "localhost:25565", "some.other.hostname.com": "localhost:1234"}}
   ```
    - `default`: Any hostnames not defined in `servers` will be routed here by default.
    - `servers`: A key-value map of hostname to `host:port`.

2. **Set the `PROXY_LISTEN_ADDR` environment variable** (optional).  
   By default, the proxy listens on `:25565`. You can change it if you want:

   ```bash
   PROXY_LISTEN_ADDR=":1234"
   ```

3. **Run the proxy**. The exact command may vary depending on your build/distribution:
   ```bash
   $ docker run -p 25565:25565 -e 'PROXY_MAPPING={"default": "localhost:25565", "servers": {"some.hostname.com": "localhost:25565", "some.other.hostname.com": "localhost:1234"}}' ghcr.io/bonnetn/minecraft-reverse-proxy:latest 
   ```

4. **Connect using the hostnames** you configured in `PROXY_MAPPING`. For example, if you set `some.hostname.com` to `localhost:25565`, your players can connect by entering `some.hostname.com` in their Minecraft client.

## Example

### Example 1: Proxy on default port (25565)

```bash
$ export PROXY_MAPPING='{
  "default": "localhost:25565",
  "servers": {
    "example.com": "192.168.1.50:25566",
    "minigames.example.com": "192.168.1.51:25565"
  }
}'
# Proxy will listen on :25565 by default
$ docker run -p 25565:25565 -e "PROXY_MAPPING=$PROXY_MAPPING" ghcr.io/bonnetn/minecraft-reverse-proxy:latest
```

- `example.com` will forward to `192.168.1.50:25566`
- `minigames.example.com` will forward to `192.168.1.51:25565`
- Any other hostname will be routed to `localhost:25565`

### Example 2: Proxy on a custom port (1234)

```bash
$ export PROXY_MAPPING='{
  "default": "localhost:25565",
  "servers": {
    "alpha.myserver.com": "localhost:25565",
    "beta.myserver.com": "localhost:25600"
  }
}'
$ export PROXY_LISTEN_ADDR=":1234"
$ docker run -p 1234:1234 -e "PROXY_MAPPING=$PROXY_MAPPING" -e "PROXY_LISTEN_ADDR=$PROXY_LISTEN_ADDR" ghcr.io/bonnetn/minecraft-reverse-proxy:latest
```

- The proxy will listen for incoming connections on port 1234.
- Hostname `alpha.myserver.com` -> `localhost:25565`
- Hostname `beta.myserver.com` -> `localhost:25600`
- Any other hostname -> `localhost:25565`

