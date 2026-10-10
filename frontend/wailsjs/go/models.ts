export namespace app {
	
	export class Appearance {
	    theme: string;
	    fontScale: number;
	
	    static createFrom(source: any = {}) {
	        return new Appearance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.fontScale = source["fontScale"];
	    }
	}
	export class ChatMessage {
	    id: string;
	    role: string;
	    content: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.status = source["status"];
	    }
	}
	export class Conversation {
	    id: string;
	    title: string;
	
	    static createFrom(source: any = {}) {
	        return new Conversation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	    }
	}
	export class SaveInput {
	    id: string;
	    name: string;
	    baseURL: string;
	    models: string[];
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
	        this.models = source["models"];
	        this.model = source["model"];
	        this.apiKey = source["apiKey"];
	        this.enabled = source["enabled"];
	    }
	}
	export class Shortcut {
	    enabled: boolean;
	    combo: string;
	
	    static createFrom(source: any = {}) {
	        return new Shortcut(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.combo = source["combo"];
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
	    models: string[];
	    activeModel: string;
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
	        this.models = source["models"];
	        this.activeModel = source["activeModel"];
	        this.credentialRef = source["credentialRef"];
	        this.hasCredential = source["hasCredential"];
	        this.enabled = source["enabled"];
	        this.insecureTransport = source["insecureTransport"];
	    }
	}

}

