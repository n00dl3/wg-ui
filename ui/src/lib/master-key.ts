import {keyToHex} from "./keygen";

const SALT_BYTE_SIZE = 16;
const IV_BYTE_SIZE = 12;

/*
Convert an array of byte values to an ArrayBuffer.
*/
function bytesToArrayBuffer(bytes: Uint8Array): ArrayBuffer {
    const bytesAsArrayBuffer = new ArrayBuffer(bytes.length);
    const bytesUint8 = new Uint8Array(bytesAsArrayBuffer);
    bytesUint8.set(bytes);
    return bytesAsArrayBuffer;
}

const getKeyMaterial = (password: ArrayBuffer): Promise<CryptoKey> => {
    return window.crypto.subtle.importKey("raw", password, {name: "PBKDF2"}, false, ["deriveBits", "deriveKey"]);
};

/*
Given some key material and some random salt
derive an AES-KW key using PBKDF2.
*/
const getUnwrappingKey = (keyMaterial: CryptoKey, salt: Uint8Array): Promise<CryptoKey> => {
    return window.crypto.subtle.deriveKey(
        {
            name: "PBKDF2",
            salt,
            iterations: 100000,
            hash: "SHA-256",
        },
        keyMaterial,
        {name: "AES-KW", length: 256},
        true,
        ["wrapKey", "unwrapKey"],
    );
};

const wrapKey = (key: CryptoKey, wrappingKey: CryptoKey): Promise<ArrayBuffer> => {
    return crypto.subtle.wrapKey("raw", key, wrappingKey, "AES-KW");
};

export const generateMasterKey = async (passphrase: ArrayBuffer): Promise<Uint8Array> => {
    const masterKey = crypto.subtle.generateKey(
        {
            name: "AES-GCM",
            length: 256,
        },
        true,
        ["encrypt", "decrypt"],
    );
    const keyMaterial = getKeyMaterial(passphrase);
    const salt = new Uint8Array(SALT_BYTE_SIZE);
    crypto.getRandomValues(salt);
    const wrappingKey = getUnwrappingKey(await keyMaterial, salt);
    const wrapped = await crypto.subtle.wrapKey("raw", await masterKey, await wrappingKey, "AES-KW");
    const output = new Uint8Array(salt.byteLength + wrapped.byteLength);
    output.set(new Uint8Array(salt), 0);
    output.set(new Uint8Array(wrapped), salt.byteLength);
    return new Uint8Array(output.buffer);
};

export const unlockMasterKey = async (passphrase: Uint8Array, wrapedKey: Uint8Array): Promise<CryptoKey> => {
    const keyMaterial = getKeyMaterial(passphrase);
    const salt = wrapedKey.slice(0, SALT_BYTE_SIZE);
    const wrapped = wrapedKey.slice(SALT_BYTE_SIZE);
    const unwrapingKey = await getUnwrappingKey(await keyMaterial, new Uint8Array(salt));
    return crypto.subtle.unwrapKey("raw", wrapped, unwrapingKey, "AES-KW", "AES-GCM", true, ["encrypt", "decrypt"]);
};

export const encrypt = async (key: CryptoKey, data: Uint8Array): Promise<Uint8Array> => {
    const iv = new Uint8Array(IV_BYTE_SIZE);
    crypto.getRandomValues(iv);
    const ciphered = await crypto.subtle.encrypt({name: "AES-GCM", iv}, key, data);
    const output = new Uint8Array(iv.byteLength + ciphered.byteLength);
    output.set(new Uint8Array(iv), 0);
    output.set(new Uint8Array(ciphered), iv.byteLength);
    return new Uint8Array(output.buffer);
};

export const decrypt = async (key: CryptoKey, data: Uint8Array): Promise<Uint8Array> => {
    const iv = data.slice(0, IV_BYTE_SIZE);
    const ciphered = data.slice(IV_BYTE_SIZE);
    return new Uint8Array(await crypto.subtle.decrypt({name: "AES-GCM", iv}, key, ciphered));
};

export class Key {}

const readKeyFromHash = (password: string) => {
    const encoder = new TextEncoder();
    unlockMasterKey(encoder.encode(password), encoder.encode(window.location.hash));
};

export class CipherException extends Error {}

export class Cipherer {
    private masterKey?: CryptoKey;
    private _wrappedKey?: Uint8Array;
    constructor() {
        console.log("Cipherer initialized");
    }
    async unlock(key: string = "", passphrase: string = "") {
        if (key != "" && passphrase != "") {
            const enc = new TextEncoder();
            this.masterKey = await unlockMasterKey(enc.encode(passphrase), enc.encode(key));
            this._wrappedKey = enc.encode(key);
        }
    }

    public get wrappedKey(): Uint8Array | undefined {
        return this._wrappedKey;
    }

    async init(passphrase: string) {
        const enc = new TextEncoder();
        const key = await generateMasterKey(enc.encode(passphrase));
        this.masterKey = await unlockMasterKey(enc.encode(passphrase), key);
        return keyToHex(key);
    }

    async encrypt(data: Uint8Array): Promise<Uint8Array> {
        if (this.masterKey) {
            return encrypt(this.masterKey, data);
        }
        throw new CipherException("Master key not initialized");
    }

    async decrypt(data: Uint8Array): Promise<Uint8Array> {
        if (this.masterKey) {
            return decrypt(this.masterKey, data);
        }
        throw new CipherException("Master key not initialized");
    }
}

const Cipher = new Cipherer();
export default Cipher;