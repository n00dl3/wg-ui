export interface IPNet {
    IP: string;
    Mask: string;
}

export function CIDRsubnetToNETIPMask(cidrmask: number) {
    let bitmask = "".padStart(cidrmask, "1").padEnd(32, "0");
    return btoa(String.fromCharCode(
        parseInt(bitmask.slice(0, 8), 2),
        parseInt(bitmask.slice(8, 16), 2),
        parseInt(bitmask.slice(16, 24), 2),
        parseInt(bitmask.slice(24, 32), 2)))
}

export function NETIPMaskToCIDRSubnet(bitmaskb64: string) {
    let bitmask = atob(bitmaskb64).split("").map((x) => x.charCodeAt(0).toString(2).padStart(8, '0')).join("");
    console.log(bitmask);
    let cidrmask = bitmask.lastIndexOf("1");
    return cidrmask == -1 ? 0 : cidrmask + 1
}

const CIDR_REGEX = /^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$/;

export function convertTextCIDRsToNETIP(allowedIPsText: string): IPNet[] | null {
    if (allowedIPsText.length == 0) {
        return null;
    }
    return allowedIPsText.split('\n').filter(cidr => {
        cidr = cidr.trim();
        return CIDR_REGEX.test(cidr);
    }).map(cidr => {
        if (cidr.indexOf('/') != -1) {
            let cidrsplit = cidr.trim().split('/');
            return {IP: cidrsplit[0], Mask: CIDRsubnetToNETIPMask(parseInt(cidrsplit[1]))}
        } else {
            return {IP: cidr, Mask: btoa("32")}
        }
    }).filter(x => !!x);
}

export function convertNETIPToTextCIDRs(netIPs: IPNet[]) {
    return netIPs.map(netip => netip.IP + "/" + NETIPMaskToCIDRSubnet(netip.Mask)).join("\n")
}
