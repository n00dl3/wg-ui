import type {IPNet} from "./ip";

export interface WGClientCreateForm {
    name?: string;
    notes?: string;
    allowedIPs: string[];
    publicKey: string;
    psk?: string;
    mtu?: number;
}

export interface WGServer {
    alowedIPs: string[];
    endpoint: string;
    publicKey: string;
}
export interface WGClient {
    ip: string;
    allowedIps: string[];
    dns: string;
    mtu: number;
    name: string;
    psk: string;
    publicKey: string;
    server: WGServer;
    keepAlive: number;
    notes: string;
    created: string;
    updated: string;
}

export interface ClientUpdateForm {
    name: string;
    notes: string;
    allowedIPs?: string[];
    dns:string
}

class APIHanlder {
    constructor(protected apiURL: string) {}

    public async create(user: string, client: WGClientCreateForm): Promise<WGClient> {
        const data = await fetch(`${this.apiURL}/api/v1/users/${user}/clients`, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(client),
        }).then((response) => {
            return response.json();
        });
        if (typeof data.Error != "undefined") {
            console.log(data.Error);
            throw new Error(data.Error);
        } else {
            console.log("New client added", data);
        }
        return data;
    }

    public async list(user: string): Promise<[string, WGClient][]> {
        let clientsUrl = `${this.apiURL}/api/v1/users/${user}/clients`;
        const clients = await fetch(clientsUrl).then((r) => r.json());
        return Object.entries(clients);
    }

    public async get(user: string, clientId: string): Promise<WGClient> {
        return await fetch(`${this.apiURL}/api/v1/users/${user}/clients/${clientId}`).then((r) => r.json());
    }

    public async update(user: string, clientId: string, client: ClientUpdateForm): Promise<WGClient> {
        return await fetch(`${this.apiURL}/api/v1/users/${user}/clients/${clientId}`, {
            method: "PUT",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(client),
        }).then((r) => r.json());
    }
    public async delete(user: string, clientId: string) {
        await fetch(`${this.apiURL}/api/v1/users/${user}/clients/${clientId}`, {
            method: "DELETE",
        });
    }

    public getQRCodeURI(user: string, clientId: string): string {
        return `${this.apiURL}/api/v1/users/${user}/clients/${clientId}?format=qrcode`;
    }

    public getConfigURI(user: string, clientId: string): string {
        return `${this.apiURL}/api/v1/users/${user}/clients/${clientId}?format=config`;
    }
}

export const client = new APIHanlder("http://localhost:8080");
export const API_URL = import.meta.env.VITE_API_URL;

export default client;
