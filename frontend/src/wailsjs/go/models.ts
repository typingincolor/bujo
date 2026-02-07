export namespace domain {
	
	export class AttentionResult {
	    Score: number;
	    Indicators: string[];
	    DaysOld: number;
	
	    static createFrom(source: any = {}) {
	        return new AttentionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Score = source["Score"];
	        this.Indicators = source["Indicators"];
	        this.DaysOld = source["DaysOld"];
	    }
	}
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
	    SortOrder: number;
	    MigrationCount: number;
	    Tags: string[];

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
	        this.SortOrder = source["SortOrder"];
	        this.MigrationCount = source["MigrationCount"];
	        this.Tags = source["Tags"];
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
	export class InsightsAction {
	    ID: number;
	    SummaryID: number;
	    ActionText: string;
	    Priority: string;
	    Status: string;
	    DueDate: string;
	    CreatedAt: string;
	    WeekStart: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightsAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.SummaryID = source["SummaryID"];
	        this.ActionText = source["ActionText"];
	        this.Priority = source["Priority"];
	        this.Status = source["Status"];
	        this.DueDate = source["DueDate"];
	        this.CreatedAt = source["CreatedAt"];
	        this.WeekStart = source["WeekStart"];
	    }
	}
	export class InsightsDecision {
	    ID: number;
	    DecisionText: string;
	    Rationale: string;
	    Participants: string;
	    ExpectedOutcomes: string;
	    DecisionDate: string;
	    SummaryID?: number;
	    CreatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightsDecision(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.DecisionText = source["DecisionText"];
	        this.Rationale = source["Rationale"];
	        this.Participants = source["Participants"];
	        this.ExpectedOutcomes = source["ExpectedOutcomes"];
	        this.DecisionDate = source["DecisionDate"];
	        this.SummaryID = source["SummaryID"];
	        this.CreatedAt = source["CreatedAt"];
	    }
	}
	export class InsightsInitiative {
	    ID: number;
	    Name: string;
	    Status: string;
	    Description: string;
	    LastUpdated: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightsInitiative(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Status = source["Status"];
	        this.Description = source["Description"];
	        this.LastUpdated = source["LastUpdated"];
	    }
	}
	export class InsightsSummary {
	    ID: number;
	    WeekStart: string;
	    WeekEnd: string;
	    SummaryText: string;
	    CreatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightsSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.WeekStart = source["WeekStart"];
	        this.WeekEnd = source["WeekEnd"];
	        this.SummaryText = source["SummaryText"];
	        this.CreatedAt = source["CreatedAt"];
	    }
	}
	export class InsightsDashboard {
	    LatestSummary?: InsightsSummary;
	    ActiveInitiatives: InsightsInitiative[];
	    HighPriorityActions: InsightsAction[];
	    RecentDecisions: InsightsDecision[];
	    DaysSinceLastSummary: number;
	    Status: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightsDashboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.LatestSummary = this.convertValues(source["LatestSummary"], InsightsSummary);
	        this.ActiveInitiatives = this.convertValues(source["ActiveInitiatives"], InsightsInitiative);
	        this.HighPriorityActions = this.convertValues(source["HighPriorityActions"], InsightsAction);
	        this.RecentDecisions = this.convertValues(source["RecentDecisions"], InsightsDecision);
	        this.DaysSinceLastSummary = source["DaysSinceLastSummary"];
	        this.Status = source["Status"];
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
	
	export class InsightsDecisionWithInitiatives {
	    ID: number;
	    DecisionText: string;
	    Rationale: string;
	    Participants: string;
	    ExpectedOutcomes: string;
	    DecisionDate: string;
	    SummaryID?: number;
	    CreatedAt: string;
	    Initiatives: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightsDecisionWithInitiatives(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.DecisionText = source["DecisionText"];
	        this.Rationale = source["Rationale"];
	        this.Participants = source["Participants"];
	        this.ExpectedOutcomes = source["ExpectedOutcomes"];
	        this.DecisionDate = source["DecisionDate"];
	        this.SummaryID = source["SummaryID"];
	        this.CreatedAt = source["CreatedAt"];
	        this.Initiatives = source["Initiatives"];
	    }
	}
	
	export class InsightsInitiativeUpdate {
	    WeekStart: string;
	    WeekEnd: string;
	    UpdateText: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightsInitiativeUpdate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.WeekStart = source["WeekStart"];
	        this.WeekEnd = source["WeekEnd"];
	        this.UpdateText = source["UpdateText"];
	    }
	}
	export class InsightsInitiativeDetail {
	    Initiative: InsightsInitiative;
	    Updates: InsightsInitiativeUpdate[];
	    PendingActions: InsightsAction[];
	    Decisions: InsightsDecision[];
	
	    static createFrom(source: any = {}) {
	        return new InsightsInitiativeDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Initiative = this.convertValues(source["Initiative"], InsightsInitiative);
	        this.Updates = this.convertValues(source["Updates"], InsightsInitiativeUpdate);
	        this.PendingActions = this.convertValues(source["PendingActions"], InsightsAction);
	        this.Decisions = this.convertValues(source["Decisions"], InsightsDecision);
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
	export class InsightsInitiativePortfolio {
	    ID: number;
	    Name: string;
	    Status: string;
	    Description: string;
	    LastUpdated: string;
	    MentionCount: number;
	    LastMentionWeek: string;
	    ActivityWeeks: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightsInitiativePortfolio(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Status = source["Status"];
	        this.Description = source["Description"];
	        this.LastUpdated = source["LastUpdated"];
	        this.MentionCount = source["MentionCount"];
	        this.LastMentionWeek = source["LastMentionWeek"];
	        this.ActivityWeeks = source["ActivityWeeks"];
	    }
	}
	
	export class InsightsInitiativeWeekUpdate {
	    InitiativeName: string;
	    UpdateText: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightsInitiativeWeekUpdate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.InitiativeName = source["InitiativeName"];
	        this.UpdateText = source["UpdateText"];
	    }
	}
	
	export class InsightsTopic {
	    ID: number;
	    SummaryID: number;
	    Topic: string;
	    Content: string;
	    Importance: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightsTopic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.SummaryID = source["SummaryID"];
	        this.Topic = source["Topic"];
	        this.Content = source["Content"];
	        this.Importance = source["Importance"];
	    }
	}
	export class InsightsTopicTimeline {
	    Topic: string;
	    Content: string;
	    Importance: string;
	    WeekStart: string;
	    WeekEnd: string;
	
	    static createFrom(source: any = {}) {
	        return new InsightsTopicTimeline(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Topic = source["Topic"];
	        this.Content = source["Content"];
	        this.Importance = source["Importance"];
	        this.WeekStart = source["WeekStart"];
	        this.WeekEnd = source["WeekEnd"];
	    }
	}
	export class InsightsWeeklyReport {
	    Summary?: InsightsSummary;
	    Topics: InsightsTopic[];
	    InitiativeUpdates: InsightsInitiativeWeekUpdate[];
	    Actions: InsightsAction[];
	
	    static createFrom(source: any = {}) {
	        return new InsightsWeeklyReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Summary = this.convertValues(source["Summary"], InsightsSummary);
	        this.Topics = this.convertValues(source["Topics"], InsightsTopic);
	        this.InitiativeUpdates = this.convertValues(source["InitiativeUpdates"], InsightsInitiativeWeekUpdate);
	        this.Actions = this.convertValues(source["Actions"], InsightsAction);
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
	
	export class ApplyResult {
	    inserted: number;
	    deleted: number;
	
	    static createFrom(source: any = {}) {
	        return new ApplyResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inserted = source["inserted"];
	        this.deleted = source["deleted"];
	    }
	}
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
	export class ResolvedDate {
	    iso: string;
	    display: string;
	
	    static createFrom(source: any = {}) {
	        return new ResolvedDate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.iso = source["iso"];
	        this.display = source["display"];
	    }
	}
	export class ValidationError {
	    lineNumber: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ValidationError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lineNumber = source["lineNumber"];
	        this.message = source["message"];
	    }
	}
	export class ValidationResult {
	    isValid: boolean;
	    errors: ValidationError[];
	
	    static createFrom(source: any = {}) {
	        return new ValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isValid = source["isValid"];
	        this.errors = this.convertValues(source["errors"], ValidationError);
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
	export class WeekSummaryDetail {
	    Summary?: domain.InsightsSummary;
	    Topics: domain.InsightsTopic[];
	
	    static createFrom(source: any = {}) {
	        return new WeekSummaryDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Summary = this.convertValues(source["Summary"], domain.InsightsSummary);
	        this.Topics = this.convertValues(source["Topics"], domain.InsightsTopic);
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

