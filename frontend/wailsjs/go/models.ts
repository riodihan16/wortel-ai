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
	    models: config.ModelEntry[];
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
	        this.models = this.convertValues(source["models"], config.ModelEntry);
	        this.model = source["model"];
	        this.apiKey = source["apiKey"];
	        this.enabled = source["enabled"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
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
	
	export class ModelEntry {
	    id: string;
	    alias: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.alias = source["alias"];
	    }
	}
	export class ProviderProfile {
	    id: string;
	    name: string;
	    baseURL: string;
	    model: string;
	    models: ModelEntry[];
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
	        this.models = this.convertValues(source["models"], ModelEntry);
	        this.activeModel = source["activeModel"];
	        this.credentialRef = source["credentialRef"];
	        this.hasCredential = source["hasCredential"];
	        this.enabled = source["enabled"];
	        this.insecureTransport = source["insecureTransport"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

