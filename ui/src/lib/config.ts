import * as QRCode from "qrcode";

interface PeerConfig {
    ipAddress: string
    name: string,
    notes: string,
    privateKey: string;
    serverPublicKey: string;
    presharedKey?: string
    serverAllowedIps: string[]
    mtu: number
    dns: string
    keepAlive: number
    serverEndpoint: string
}


export const generateConfig = ({ ipAddress, privateKey, dns,
    mtu, serverPublicKey, serverAllowedIps, serverEndpoint,
    keepAlive, presharedKey }: PeerConfig): string => {
    return `[Interface]
Address=${ipAddress}
PrivateKey=${privateKey}
DNS=${dns}
MTU=${mtu}

[Peer]
PublicKey=${serverPublicKey}
AllowedIPs=${serverAllowedIps.join(',')}
Endpoint=${serverEndpoint}
PersistentKeepalive=${keepAlive}
PresharedKey=${presharedKey}`
}


export const getQrcodeConfig=(cfg:PeerConfig):Promise<string>=>{
    return QRCode.toDataURL(generateConfig(cfg))
}