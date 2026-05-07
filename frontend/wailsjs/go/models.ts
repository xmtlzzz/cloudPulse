export namespace core {
	
	export class AlertInfo {
	    platform: string;
	    resource: string;
	    percentage: number;
	    // Go type: time
	    time: any;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new AlertInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.resource = source["resource"];
	        this.percentage = source["percentage"];
	        this.time = this.convertValues(source["time"], null);
	        this.status = source["status"];
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
	export class AppSettings {
	    theme: string;
	    syncInterval: number;
	    alertThreshold: number;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.syncInterval = source["syncInterval"];
	        this.alertThreshold = source["alertThreshold"];
	    }
	}
	export class AuthField {
	    id: string;
	    name: string;
	    description: string;
	    type: string;
	    required: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AuthField(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.type = source["type"];
	        this.required = source["required"];
	    }
	}
	export class AuthConfig {
	    fields: AuthField[];
	
	    static createFrom(source: any = {}) {
	        return new AuthConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fields = this.convertValues(source["fields"], AuthField);
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
	
	export class ResourceUsage {
	    categoryId: string;
	    name: string;
	    used: number;
	    limit: number;
	    unit: string;
	    // Go type: time
	    resetDate?: any;
	
	    static createFrom(source: any = {}) {
	        return new ResourceUsage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.categoryId = source["categoryId"];
	        this.name = source["name"];
	        this.used = source["used"];
	        this.limit = source["limit"];
	        this.unit = source["unit"];
	        this.resetDate = this.convertValues(source["resetDate"], null);
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
	export class UsageData {
	    platform: string;
	    accountName: string;
	    plan: string;
	    resources: ResourceUsage[];
	    // Go type: time
	    lastUpdated: any;
	    status: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new UsageData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.accountName = source["accountName"];
	        this.plan = source["plan"];
	        this.resources = this.convertValues(source["resources"], ResourceUsage);
	        this.lastUpdated = this.convertValues(source["lastUpdated"], null);
	        this.status = source["status"];
	        this.error = source["error"];
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
	export class DashboardData {
	    services: UsageData[];
	    alerts: AlertInfo[];
	    // Go type: time
	    lastSync: any;
	    connectedCount: number;
	    healthyCount: number;
	    warnCount: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.services = this.convertValues(source["services"], UsageData);
	        this.alerts = this.convertValues(source["alerts"], AlertInfo);
	        this.lastSync = this.convertValues(source["lastSync"], null);
	        this.connectedCount = source["connectedCount"];
	        this.healthyCount = source["healthyCount"];
	        this.warnCount = source["warnCount"];
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

export namespace custom {
	
	export class ResourceMapping {
	    name: string;
	    path: string;
	    limitPath: string;
	    unit: string;
	    defaultLimit: number;
	
	    static createFrom(source: any = {}) {
	        return new ResourceMapping(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.limitPath = source["limitPath"];
	        this.unit = source["unit"];
	        this.defaultLimit = source["defaultLimit"];
	    }
	}
	export class CustomServiceConfig {
	    name: string;
	    description: string;
	    baseUrl: string;
	    headers: Record<string, string>;
	    authType: string;
	    authKey: string;
	    resources: ResourceMapping[];
	
	    static createFrom(source: any = {}) {
	        return new CustomServiceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.baseUrl = source["baseUrl"];
	        this.headers = source["headers"];
	        this.authType = source["authType"];
	        this.authKey = source["authKey"];
	        this.resources = this.convertValues(source["resources"], ResourceMapping);
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

