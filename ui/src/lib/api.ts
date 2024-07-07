import type {IPNet} from "./ip";

export interface WGClientCreateForm {
    Name: string;
    Notes: string;
    generatePSK: boolean;
}

export interface WGClient {
    Name: string;
    PrivateKey: string;
    PublicKey: string;
    PresharedKey: string;
    IP: string;
    AllowedIPs: IPNet[]|null,
    MTU: number;
    Notes: string;
    Created: string;
    Modified: string;
}

export interface ClientUpdateForm {
    Name: string;
    Notes: string;
    AllowedIPs: IPNet[] | null;
}


class APIHanlder {

    constructor(protected apiURL: string) {
    }

    public async create(user: string, client: WGClientCreateForm) {
        const data = await fetch(`${this.apiURL}/api/v1/users/${user}/clients`, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(client),
        }).then(response => {
            return response.json();
        })
        if (typeof data.Error != "undefined") {
            console.log(data.Error);
            alert(data.Error);
        } else {
            console.log("New client added", data);
        }
    }

    public async list(user: string): Promise<[string, WGClient][]> {
        let clientsUrl = `${this.apiURL}/api/v1/users/${user}/clients`;
        const clients = await fetch(clientsUrl).then(r => r.json())
        return Object.entries(clients);
    }

    public async get(user: string, clientId: string): Promise<WGClient> {
        return await fetch(`${this.apiURL}/api/v1/users/${user}/clients/${clientId}`).then(r => r.json());
    }

    public async update(user: string, clientId: string, client: ClientUpdateForm): Promise<WGClient> {
        return await fetch(`${this.apiURL}/api/v1/users/${user}/clients/${clientId}`, {
            method: "PUT",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(client),
        }).then(r => r.json());
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

