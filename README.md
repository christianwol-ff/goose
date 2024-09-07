
<h2 align="center">
# Decentralized Tunnel Network - Goose

[![Build](https://github.com/nickjfree/goose/actions/workflows/build.yml/badge.svg)](https://github.com/nickjfree/goose/actions/workflows/build.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/nickjfree/goose)](https://goreportcard.com/report/github.com/nickjfree/goose)

</h2>



## Features

- **Config-Free Node Discovery**: Eliminates the need for manual configuration by automatically discovering peers in the network. It uses the libp2p network and is bootstrapped via the IPFS network, making the setup hassle-free.

- **Protocol Support**: Offers flexibility by supporting multiple protocols, including QUIC and WireGuard. This allows users to choose the protocol that best suits their needs.

- **Virtual Private Network**: Creates a virtual network interface named `goose`, enabling secure and private communication channels over the internet.

- **Fake-IP**:  Utilizes the `fake-ip` method to selectively route traffic either through the secure tunnel interface or directly to the real network interface. This feature allows for more granular control over traffic routing. Users can write custom scripts to handle the selection of routing, making it highly customizable.


## Usage [🤖](https://chat.openai.com/g/g-CMQzJ1mTq-goose-grid-commander)

Run the following command to see the available options:

```bash
goose -h
Usage of goose:
  -e string

        comma separated remote endpoints.
        eg. ipfs/QmVCVa7RfutQDjvUYTejMyVLMMF5xYAM1mEddDVwMmdLf4,ipfs/QmYXWTQ1jTZ3ZEXssCyBHMh4H4HqLPez5dhpqkZbSJjh7r

  -f string
        forward networks, comma separated CIDRs
  -g string
        geoip db file
  -l string

        virtual ip address to use in CIDR format.
        local ipv4 address to set on the tunnel interface.
         (default "192.168.32.166/24")
  -n string
        namespace
  -name string
        domain name to use, namespace must be set
  -p string
        fake ip range
  -r string
        rule script
  -wg string
        wireguard config file
```


## Examples

### Simple Connection

1. On Computer A, run:

```bash
    goose -n my-network -name a
```

2. On Computer B, run:

```bash
    goose -n my-network -name b
```

3. After a few minutes, they will connect. You can ping B from A using:

```bash
ping a.my-network

64 bytes from a.goose.my-network(192.168.0.4): icmp_seq=1 ttl=63 time=188 ms
64 bytes from a.goose.my-network(192.168.0.4): icmp_seq=2 ttl=63 time=206 ms
64 bytes from a.goose.my-network(192.168.0.4): icmp_seq=3 ttl=63 time=748 ms
64 bytes from a.goose.my-network(192.168.0.4): icmp_seq=4 ttl=63 time=562 ms
```

### Network Forwarding

1. Assume Computer A is connected to a private network `10.1.1.0/24`.

2. On Computer A, run:

```bash
    goose -n my-network -name a -f 10.1.1.0/24
```

3. On Computer B, run:

```bash
    goose -n my-network -name b
```

4. Now you can access any host in `10.1.1.0/24` from Computer B using:

```bash
ping 10.1.1.1

64 bytes from 10.1.1.1: icmp_seq=1 ttl=63 time=188 ms
64 bytes from 10.1.1.1: icmp_seq=2 ttl=63 time=206 ms
64 bytes from 10.1.1.1: icmp_seq=3 ttl=63 time=748 ms
64 bytes from 10.1.1.1: icmp_seq=4 ttl=63 time=562 ms
```

### Fake-IP Example

1. On Computer A, run:

```bash
    goose -n my-network -name a -f 0.0.0.0/0
```

2. On Computer B:

####  Custom Script for Routing (Optional)

Use `rule.js` to define custom routing rules.

The custom script must define a `matchDomain(domain)` function. Any traffic that matches the criteria set in this function will bypass the tunnel and be routed directly to the real network interface.

The scripts should be written in ES5

Here's an example:

```javascript
// rule.js
var filters = ['baidu', 'shifen', 'csdn', 'qq', 'libp2p'];
var filterRegions = ['CN'];

function isIPv4(str) {
  var ipv4Regex = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
  return ipv4Regex.test(str);
}

// Define the main function to match a domain
function matchDomain(domain) {
  if (isIPv4(domain)) {
    var country = getCountry(domain); 
    return filterRegions.indexOf(country) !== -1
  }
  else if (filters.some(function(name) {
    return domain.indexOf(name) !== -1;
  })) {
    return true;
  }
  return false;
}
```
Run the following command to apply the custom rules:

```bash
goose -n my-network -name b -g geoip-country.mmdb -r rule.js -p 11.0.0.0/16
```

Explanation: This command applies the custom routing rules defined in rule.js and sets up a fake-ip range of 11.0.0.0/16.


Testing

```bash
ping www.google.com

PING www.google.com (11.0.0.133) 56(84) bytes of data.
64 bytes from 10.0.0.133 (10.0.0.133): icmp_seq=1 ttl=59 time=188 ms
64 bytes from 10.0.0.133 (10.0.0.133): icmp_seq=2 ttl=59 time=189 ms
64 bytes from 10.0.0.133 (10.0.0.133): icmp_seq=3 ttl=59 time=188 ms
64 bytes from 10.0.0.133 (10.0.0.133): icmp_seq=4 ttl=59 time=188 ms

ping www.baidu.com

PING www.wshifen.com (104.193.88.123) 56(84) bytes of data.
64 bytes from 104.193.88.123 (104.193.88.123): icmp_seq=1 ttl=50 time=150 ms
64 bytes from 104.193.88.123 (104.193.88.123): icmp_seq=2 ttl=50 time=149 ms
64 bytes from 104.193.88.123 (104.193.88.123): icmp_seq=3 ttl=50 time=149 ms
```

### WireGuard Example

WireGuard is a modern, secure, and fast VPN tunnel that aims to be easy to use and lean.

#### Example WireGuard Config File

Below is an example of a WireGuard configuration file that can be used with Goose:

```bash
[Interface]
PrivateKey = mIz7fpuVMc4p1S3e3D4sifkq1fGtgzRJs/kgcuYARWE=
ListenPort = 51820

[Peer]  
PublicKey = CdjruGQqzRC5zUUQEPNjXRPlbmj5t/C0VzF+g93wGkM=
AllowedIPs = 10.0.0.1/32
PersistentKeepalive = 25

PublicKey = x0BPthZpWvmt+KagQgX1zdCQtAHi1Rv6PhcHkOb1cjA=
AllowedIPs = 10.0.0.2/32
PersistentKeepalive = 25

PublicKey = CNx+uklxUet6JQASvh315s1zKqsXh8n1sm3PYUNgeiU=
AllowedIPs = 10.0.0.3/32
PersistentKeepalive = 25
```

#### Running the WireGuard Command

To integrate WireGuard with Goose, run the following command:

```bash
goose -n my-network -name a -wg /etc/wg.conf
```

This command does the following:

- `-n my-network`: Specifies the virtual network name as `my-network`.
- `-name a`: Sets the node name to `a`.
- `-wg /etc/wg.conf`: Points to the WireGuard configuration file located at `/etc/wg.conf`.

#### Connecting to the Virtual Network

After running this command, you can connect to the virtual `my-network` using any WireGuard client implementation.


Проєкт **Goose** на GitHub (https://github.com/nickjfree/goose) є програмою для тунелювання даних через різні мережі. Ось детальний опис технології, її можливостей та приклади застосування:

### Опис технології
**Goose** — це багатофункціональний інструмент для тунелювання мережевого трафіку через альтернативні мережі з метою обійти мережеві обмеження або покращити захист даних під час передачі. Основна мета програми полягає в тому, щоб створити надійний канал для передачі інформації, що може бути корисним у ситуаціях, коли необхідно приховати або зашифрувати передавання даних.

### Основні можливості:
1. **Створення тунелів**: Goose дозволяє створювати тунелі для мережевого трафіку через різні мережеві шари, включаючи TCP, UDP, HTTP/HTTPS.
2. **Підтримка обфускації трафіку**: Щоб уникати виявлення або блокування провайдерами, Goose використовує різні техніки обфускації (приховування) трафіку.
3. **Підтримка безпечного підключення**: Шифрування трафіку дозволяє захищати передані дані від перехоплення третіми сторонами.
4. **Гнучка конфігурація**: Інструмент підтримує налаштування для різних типів трафіку, що дозволяє підлаштувати його під конкретні вимоги користувача.
5. **Підтримка різних платформ**: Goose сумісний з різними операційними системами, включаючи Linux, Windows та macOS.

### Приклади застосування:
1. **Обхід цензури та обмежень**: Goose може використовуватися для обходу мережевих блокувань в країнах, де доступ до певних ресурсів обмежений.
   
2. **Захист приватності**: Використовуючи тунелі та шифрування, користувачі можуть захищати свої дані під час користування незахищеними мережами, наприклад, громадським Wi-Fi.

3. **Оптимізація мережевого трафіку**: В деяких випадках, коли стандартні маршрути мережевого трафіку занадто завантажені або небезпечні, тунелювання через альтернативні маршрути може допомогти підвищити швидкість та безпеку передачі даних.

4. **Корпоративне застосування**: В корпоративних середовищах тунелювання трафіку через безпечні канали дозволяє забезпечити захищений доступ до внутрішніх ресурсів компанії з будь-якого місця.

### Приклади використання:
- **Активісти або журналісти** можуть використовувати Goose для доступу до інформації в умовах обмеженої свободи слова.
- **IT-фахівці та системні адміністратори** застосовують інструменти тунелювання для організації віддаленого доступу до корпоративних мереж або для безпечної передачі конфіденційних даних.
  
### Наглядні кейси:
- **Обхід блокувань в країнах з інтернет-цензурою**: У деяких країнах існують серйозні обмеження на доступ до глобальної мережі. Goose може допомогти створити захищені канали для доступу до блокованих веб-сайтів і сервісів.
  
- **Забезпечення анонімності в інтернеті**: Використання інструменту допоможе приховати дії користувачів у мережі та захистити їхню особисту інформацію.

### Висновок:
Goose — це потужний інструмент для тунелювання трафіку, що дозволяє вирішувати різні завдання з обфускації, захисту та обходу мережевих обмежень. Цей інструмент особливо корисний в умовах підвищеної цензури або для тих, хто потребує надійного захисту своїх даних під час роботи в незахищених мережах.

