export namespace app {
	
	export class SaveInput {
	    id: string;
	    name: string;
	    baseURL: string;
	    model: string;
	    apiKey: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SaveInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.baseURL = source["baseURL"];
	        this.model = source["model"];
	        this.apiKey = source["apiKey"];
	        this.enabled = source["enabled"];
	    }
	}
	export class TestResult {
	    ok: boolean;
	    code?: string;
	    message: string;
	    retryable?: boolean;
	    statusCode?: number;
	    retryAfterSec?: number;
	    insecure?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.retryable = source["retryable"];
	        this.statusCode = source["statusCode"];
	        this.retryAfterSec = source["retryAfterSec"];
	        this.insecure = source["insecure"];
	    }
	}

}

export namespace config {
	
	export class ProviderProfile {
	    id: string;
	    name: string;
	    baseURL: string;
	    model: string;
	    credentialRef: string;
	    hasCredential: boolean;
	    enabled: boolean;
	    insecureTransport?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProviderProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.baseURL = source["baseURL"];
	        this.model = source["model"];
	        this.credentialRef = source["credentialRef"];
	        this.hasCredential = source["hasCredential"];
	        this.enabled = source["enabled"];
	        this.insecureTransport = source["insecureTransport"];
	    }
	}

}

