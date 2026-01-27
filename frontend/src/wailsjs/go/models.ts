export namespace domain {
	
	export class Entry {
	    ID: number;
	    EntityID: string;
	    Type: string;
	    Content: string;
	    Priority: string;
	    ParentID?: number;
	    ParentEntityID?: string;
	    Depth: number;
	    Location?: string;
	    ScheduledDate?: time.Time;
	    CreatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.EntityID = source["EntityID"];
	        this.Type = source["Type"];
	        this.Content = source["Content"];
	        this.Priority = source["Priority"];
	        this.ParentID = source["ParentID"];
	        this.ParentEntityID = source["ParentEntityID"];
	        this.Depth = source["Depth"];
	        this.Location = source["Location"];
	        this.ScheduledDate = this.convertValues(source["ScheduledDate"], time.Time);
	        this.CreatedAt = this.convertValues(source["CreatedAt"], time.Time);
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
	export class Goal {
	    ID: number;
	    EntityID: string;
	    Content: string;
	    Month: time.Time;
	    Status: string;
	    MigratedTo?: time.Time;
	    CreatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Goal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.EntityID = source["EntityID"];
	        this.Content = source["Content"];
	        this.Month = this.convertValues(source["Month"], time.Time);
	        this.Status = source["Status"];
	        this.MigratedTo = this.convertValues(source["MigratedTo"], time.Time);
	        this.CreatedAt = this.convertValues(source["CreatedAt"], time.Time);
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
	export class ListItem {
	    RowID: number;
	    EntityID: string;
	    Version: number;
	    ValidFrom: time.Time;
	    ValidTo?: time.Time;
	    OpType: string;
	    ListEntityID: string;
	    Type: string;
	    Content: string;
	    CreatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new ListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.RowID = source["RowID"];
	        this.EntityID = source["EntityID"];
	        this.Version = source["Version"];
	        this.ValidFrom = this.convertValues(source["ValidFrom"], time.Time);
	        this.ValidTo = this.convertValues(source["ValidTo"], time.Time);
	        this.OpType = source["OpType"];
	        this.ListEntityID = source["ListEntityID"];
	        this.Type = source["Type"];
	        this.Content = source["Content"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], time.Time);
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

export namespace service {
	
	export class DayEntries {
	    Date: time.Time;
	    Location?: string;
	    Mood?: string;
	    Weather?: string;
	    Entries: domain.Entry[];
	
	    static createFrom(source: any = {}) {
	        return new DayEntries(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Date = this.convertValues(source["Date"], time.Time);
	        this.Location = source["Location"];
	        this.Mood = source["Mood"];
	        this.Weather = source["Weather"];
	        this.Entries = this.convertValues(source["Entries"], domain.Entry);
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
	export class DayStatus {
	    Date: time.Time;
	    Completed: boolean;
	    Count: number;
	
	    static createFrom(source: any = {}) {
	        return new DayStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Date = this.convertValues(source["Date"], time.Time);
	        this.Completed = source["Completed"];
	        this.Count = source["Count"];
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
	export class HabitStatus {
	    ID: number;
	    Name: string;
	    GoalPerDay: number;
	    GoalPerWeek: number;
	    GoalPerMonth: number;
	    CurrentStreak: number;
	    CompletionPercent: number;
	    WeeklyProgress: number;
	    MonthlyProgress: number;
	    TodayCount: number;
	    DayHistory: DayStatus[];
	
	    static createFrom(source: any = {}) {
	        return new HabitStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.GoalPerDay = source["GoalPerDay"];
	        this.GoalPerWeek = source["GoalPerWeek"];
	        this.GoalPerMonth = source["GoalPerMonth"];
	        this.CurrentStreak = source["CurrentStreak"];
	        this.CompletionPercent = source["CompletionPercent"];
	        this.WeeklyProgress = source["WeeklyProgress"];
	        this.MonthlyProgress = source["MonthlyProgress"];
	        this.TodayCount = source["TodayCount"];
	        this.DayHistory = this.convertValues(source["DayHistory"], DayStatus);
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
	export class TrackerStatus {
	    Habits: HabitStatus[];
	
	    static createFrom(source: any = {}) {
	        return new TrackerStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Habits = this.convertValues(source["Habits"], HabitStatus);
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

export namespace time {
	
	export class Time {
	
	
	    static createFrom(source: any = {}) {
	        return new Time(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace wails {
	
	export class ListWithItems {
	    ID: number;
	    Name: string;
	    Items: domain.ListItem[];
	
	    static createFrom(source: any = {}) {
	        return new ListWithItems(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Items = this.convertValues(source["Items"], domain.ListItem);
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

