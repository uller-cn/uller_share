export namespace common {
	
	export class Share {
	    shareId: number;
	    title: string;
	    localPath: string;
	    ext: string;
	    size: number;
	    expireTime: number;
	    ip: string;
	    // Go type: time
	    createTime: any;
	    // Go type: time
	    updateTime: any;
	
	    static createFrom(source: any = {}) {
	        return new Share(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shareId = source["shareId"];
	        this.title = source["title"];
	        this.localPath = source["localPath"];
	        this.ext = source["ext"];
	        this.size = source["size"];
	        this.expireTime = source["expireTime"];
	        this.ip = source["ip"];
	        this.createTime = this.convertValues(source["createTime"], null);
	        this.updateTime = this.convertValues(source["updateTime"], null);
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
	export class DownLoadHistory {
	    historyId: number;
	    title: string;
	    localPath: string;
	    ip: string;
	    share: Share;
	    ext: string;
	    size: number;
	    finish: number;
	    status: number;
	    // Go type: time
	    createTime: any;
	    // Go type: time
	    updateTime: any;
	
	    static createFrom(source: any = {}) {
	        return new DownLoadHistory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.historyId = source["historyId"];
	        this.title = source["title"];
	        this.localPath = source["localPath"];
	        this.ip = source["ip"];
	        this.share = this.convertValues(source["share"], Share);
	        this.ext = source["ext"];
	        this.size = source["size"];
	        this.finish = source["finish"];
	        this.status = source["status"];
	        this.createTime = this.convertValues(source["createTime"], null);
	        this.updateTime = this.convertValues(source["updateTime"], null);
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
	export class DownLoadHistoryList {
	    downLoadHistory: DownLoadHistory[];
	    total?: number;
	
	    static createFrom(source: any = {}) {
	        return new DownLoadHistoryList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.downLoadHistory = this.convertValues(source["downLoadHistory"], DownLoadHistory);
	        this.total = source["total"];
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

