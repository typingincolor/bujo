import { useState, useEffect } from 'react';
import { Settings, Palette, Database, Info, Mail, Tablet } from 'lucide-react';
import { useSettings } from '../../contexts/SettingsContext';
import type { Theme, DefaultView } from '../../types/settings';
import { GetVersion, IsRemarkableRegistered, RegisterRemarkableDevice } from '@/wailsjs/go/wails/App';
import { BrowserOpenURL } from '@/wailsjs/runtime/runtime';

interface SettingsViewProps {
  onRemarkableRegistered?: () => void
}

export function SettingsView({ onRemarkableRegistered }: SettingsViewProps) {
  const { theme, setTheme, defaultView, setDefaultView } = useSettings();
  const [version, setVersion] = useState<string>('Loading...');
  const [remarkableRegistered, setRemarkableRegistered] = useState<boolean | null>(null);
  const [remarkableCode, setRemarkableCode] = useState('');
  const [remarkableRegistering, setRemarkableRegistering] = useState(false);
  const [remarkableError, setRemarkableError] = useState<string | null>(null);
  const [remarkableSuccess, setRemarkableSuccess] = useState(false);

  useEffect(() => {
    GetVersion().then(setVersion).catch(() => setVersion('Unknown'));
    IsRemarkableRegistered().then(setRemarkableRegistered);
  }, []);

  async function handleRemarkableRegister() {
    if (!remarkableCode.trim()) return;
    setRemarkableRegistering(true);
    setRemarkableError(null);
    setRemarkableSuccess(false);
    try {
      await RegisterRemarkableDevice(remarkableCode.trim());
      setRemarkableRegistered(true);
      setRemarkableSuccess(true);
      setRemarkableCode('');
      onRemarkableRegistered?.();
    } catch (err) {
      setRemarkableError(String(err));
    } finally {
      setRemarkableRegistering(false);
    }
  }
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
            <ThemeSelector currentTheme={theme} onThemeChange={setTheme} />
          </SettingRow>
          <SettingRow
            label="Default View"
            description="The view shown when you open the app"
          >
            <DefaultViewSelector currentView={defaultView} onViewChange={setDefaultView} />
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
            <span className="text-xs text-muted-foreground">
              ~/.bujo/bujo.db
            </span>
          </SettingRow>
        </div>
      </section>

      {/* Integrations Section */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Mail className="w-4 h-4 text-muted-foreground" />
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
            Integrations
          </h3>
        </div>
        <div className="rounded-lg border border-border bg-card p-4 space-y-4">
          <SettingRow
            label="Gmail Bookmarklet"
            description="Capture emails as tasks directly from Gmail"
          >
            <button
              onClick={() => BrowserOpenURL('http://127.0.0.1:8743/install')}
              className="text-sm text-primary hover:underline"
            >
              Install
            </button>
          </SettingRow>
        </div>
      </section>

      {/* reMarkable Section */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Tablet className="w-4 h-4 text-muted-foreground" />
          <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
            reMarkable
          </h3>
        </div>
        <div className="rounded-lg border border-border bg-card p-4 space-y-4">
          <SettingRow
            label="Device Registration"
            description={remarkableRegistered ? 'Your reMarkable tablet is connected' : 'Connect your reMarkable tablet to import handwritten notes'}
          >
            {remarkableRegistered === null ? (
              <span className="text-xs text-muted-foreground">Checking...</span>
            ) : remarkableRegistered ? (
              <span className="text-sm text-green-600">Connected</span>
            ) : (
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  value={remarkableCode}
                  onChange={e => setRemarkableCode(e.target.value)}
                  placeholder="One-time code"
                  className="px-2 py-1 text-sm bg-background border border-border rounded w-32"
                  onKeyDown={e => e.key === 'Enter' && handleRemarkableRegister()}
                />
                <button
                  onClick={handleRemarkableRegister}
                  disabled={remarkableRegistering || !remarkableCode.trim()}
                  className="px-3 py-1 text-sm bg-primary text-primary-foreground rounded disabled:opacity-50"
                >
                  {remarkableRegistering ? 'Registering...' : 'Register'}
                </button>
              </div>
            )}
          </SettingRow>
          {remarkableError && (
            <p className="text-xs text-destructive">{remarkableError}</p>
          )}
          {remarkableSuccess && (
            <p className="text-xs text-green-600">Device registered successfully!</p>
          )}
          {!remarkableRegistered && remarkableRegistered !== null && (
            <p className="text-xs text-muted-foreground">
              Get a one-time code from{' '}
              <button
                onClick={() => BrowserOpenURL('https://my.remarkable.com/device/browser/connect')}
                className="text-primary hover:underline"
              >
                my.remarkable.com
              </button>
            </p>
          )}
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
            <span className="text-sm text-muted-foreground">{version}</span>
          </SettingRow>
          <SettingRow
            label="bujo"
            description="Your digital bullet journal"
          >
            <button
              onClick={() => BrowserOpenURL('https://github.com/typingincolor/bujo')}
              className="text-sm text-primary hover:underline"
            >
              GitHub
            </button>
          </SettingRow>
          <SettingRow
            label="Support"
            description="Report bugs or request features"
          >
            <button
              onClick={() => BrowserOpenURL('https://github.com/typingincolor/bujo/issues')}
              className="text-sm text-primary hover:underline"
            >
              GitHub Issues
            </button>
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
    <div className="flex items-center justify-between flex-wrap gap-2">
      <div className="min-w-0">
        <p className="text-sm font-medium">{label}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
      {children}
    </div>
  );
}

interface ThemeSelectorProps {
  currentTheme: Theme;
  onThemeChange: (theme: Theme) => void;
}

function ThemeSelector({ currentTheme, onThemeChange }: ThemeSelectorProps) {
  const themes: { value: Theme; label: string }[] = [
    { value: 'light', label: 'Light' },
    { value: 'dark', label: 'Dark' },
    { value: 'system', label: 'System' },
  ];

  return (
    <div className="flex gap-2">
      {themes.map(({ value, label }) => (
        <button
          key={value}
          onClick={() => onThemeChange(value)}
          className={`px-3 py-1 text-sm rounded transition-colors ${
            currentTheme === value
              ? 'bg-primary text-primary-foreground'
              : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
          }`}
        >
          {label}
        </button>
      ))}
    </div>
  );
}

interface DefaultViewSelectorProps {
  currentView: DefaultView;
  onViewChange: (view: DefaultView) => void;
}

function DefaultViewSelector({ currentView, onViewChange }: DefaultViewSelectorProps) {
  const views: { value: DefaultView; label: string }[] = [
    { value: 'today', label: 'Today' },
    { value: 'week', label: 'Week' },
    { value: 'search', label: 'Search' },
  ];

  return (
    <div className="flex gap-2">
      {views.map(({ value, label }) => (
        <button
          key={value}
          onClick={() => onViewChange(value)}
          className={`px-3 py-1 text-sm rounded transition-colors ${
            currentView === value
              ? 'bg-primary text-primary-foreground'
              : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
          }`}
        >
          {label}
        </button>
      ))}
    </div>
  );
}
