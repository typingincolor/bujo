import { Settings, Palette, Database, Info } from 'lucide-react';

export function SettingsView() {
  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center gap-2">
        <Settings className="w-5 h-5 text-primary" />
        <h2 className="font-display text-xl font-semibold">Settings</h2>
      </div>

      {/* Appearance Section */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Palette className="w-4 h-4 text-muted-foreground" />
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
            Appearance
          </h3>
        </div>
        <div className="rounded-lg border border-border bg-card p-4 space-y-4">
          <SettingRow
            label="Theme"
            description="Choose your preferred color theme"
          >
            <span className="text-sm text-muted-foreground">Dark</span>
          </SettingRow>
          <SettingRow
            label="Default View"
            description="The view shown when you open the app"
          >
            <span className="text-sm text-muted-foreground">Today</span>
          </SettingRow>
        </div>
      </section>

      {/* Data Section */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Database className="w-4 h-4 text-muted-foreground" />
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
            Data
          </h3>
        </div>
        <div className="rounded-lg border border-border bg-card p-4 space-y-4">
          <SettingRow
            label="Database Location"
            description="Where your journal data is stored"
          >
            <span className="text-xs text-muted-foreground font-mono">
              ~/.bujo/bujo.db
            </span>
          </SettingRow>
        </div>
      </section>

      {/* About Section */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Info className="w-4 h-4 text-muted-foreground" />
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
            About
          </h3>
        </div>
        <div className="rounded-lg border border-border bg-card p-4 space-y-4">
          <SettingRow
            label="Version"
            description="Current application version"
          >
            <span className="text-sm text-muted-foreground">1.0.0</span>
          </SettingRow>
          <SettingRow
            label="bujo"
            description="Your digital bullet journal"
          >
            <a
              href="https://github.com/typingincolor/bujo"
              className="text-sm text-primary hover:underline"
              target="_blank"
              rel="noopener noreferrer"
            >
              GitHub
            </a>
          </SettingRow>
        </div>
      </section>
    </div>
  );
}

interface SettingRowProps {
  label: string;
  description: string;
  children: React.ReactNode;
}

function SettingRow({ label, description, children }: SettingRowProps) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <p className="text-sm font-medium">{label}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
      {children}
    </div>
  );
}
