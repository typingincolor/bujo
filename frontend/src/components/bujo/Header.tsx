import { useState, useEffect, useRef } from 'react';
import { format } from 'date-fns';
import { Calendar, FileEdit, Smile, Cloud, MapPin } from 'lucide-react';
import { cn } from '@/lib/utils';
import { SetMood, SetWeather, SetLocation, GetLocationHistory } from '@/wailsjs/go/wails/App';

const MOOD_OPTIONS = [
  { emoji: '😊', value: 'happy' },
  { emoji: '😐', value: 'neutral' },
  { emoji: '😢', value: 'sad' },
  { emoji: '😤', value: 'frustrated' },
  { emoji: '😴', value: 'tired' },
  { emoji: '🤒', value: 'sick' },
  { emoji: '😰', value: 'anxious' },
  { emoji: '🤗', value: 'grateful' },
] as const;

const WEATHER_OPTIONS = [
  { emoji: '☀️', value: 'sunny' },
  { emoji: '🌤️', value: 'partly-cloudy' },
  { emoji: '☁️', value: 'cloudy' },
  { emoji: '🌧️', value: 'rainy' },
  { emoji: '⛈️', value: 'stormy' },
  { emoji: '❄️', value: 'snowy' },
] as const;

const LOCATION_OPTIONS = [
  { emoji: '🏠', value: 'home' },
  { emoji: '🏢', value: 'office' },
  { emoji: '☕', value: 'cafe' },
  { emoji: '📚', value: 'library' },
  { emoji: '✈️', value: 'travel' },
] as const;

type MoodValue = typeof MOOD_OPTIONS[number]['value'];
type WeatherValue = typeof WEATHER_OPTIONS[number]['value'];

interface HeaderProps {
  title: string;
  onCapture?: () => void;
  currentMood?: string;
  currentWeather?: string;
  currentLocation?: string;
  onMoodChanged?: () => void;
  onWeatherChanged?: () => void;
  onLocationChanged?: () => void;
}

export function Header({
  title,
  onCapture,
  currentMood,
  currentWeather,
  currentLocation,
  onMoodChanged,
  onWeatherChanged,
  onLocationChanged,
}: HeaderProps) {
  const today = new Date();
  const [showMoodPicker, setShowMoodPicker] = useState(false);
  const [showWeatherPicker, setShowWeatherPicker] = useState(false);
  const [showLocationPicker, setShowLocationPicker] = useState(false);
  const [locationInput, setLocationInput] = useState('');
  const [locationHistory, setLocationHistory] = useState<string[]>([]);
  const moodPickerRef = useRef<HTMLDivElement>(null);
  const weatherPickerRef = useRef<HTMLDivElement>(null);
  const locationPickerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (moodPickerRef.current && !moodPickerRef.current.contains(e.target as Node)) {
        setShowMoodPicker(false);
      }
      if (weatherPickerRef.current && !weatherPickerRef.current.contains(e.target as Node)) {
        setShowWeatherPicker(false);
      }
      if (locationPickerRef.current && !locationPickerRef.current.contains(e.target as Node)) {
        setShowLocationPicker(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  useEffect(() => {
    if (showLocationPicker) {
      GetLocationHistory().then(setLocationHistory).catch(() => setLocationHistory([]));
    }
  }, [showLocationPicker]);

  const handleMoodSelect = async (mood: MoodValue) => {
    await SetMood(today.toISOString(), mood);
    setShowMoodPicker(false);
    onMoodChanged?.();
  };

  const handleWeatherSelect = async (weather: WeatherValue) => {
    await SetWeather(today.toISOString(), weather);
    setShowWeatherPicker(false);
    onWeatherChanged?.();
  };

  const handleLocationSelect = async (location: string) => {
    if (!location.trim()) return;
    await SetLocation(today.toISOString(), location);
    setShowLocationPicker(false);
    setLocationInput('');
    onLocationChanged?.();
  };

  const handleLocationKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleLocationSelect(locationInput);
    }
  };

  const getMoodEmoji = (mood: string) => {
    return MOOD_OPTIONS.find(m => m.value === mood)?.emoji;
  };

  const getWeatherEmoji = (weather: string) => {
    return WEATHER_OPTIONS.find(w => w.value === weather)?.emoji;
  };

  const getLocationEmoji = (location: string) => {
    return LOCATION_OPTIONS.find(l => l.value === location)?.emoji;
  };

  return (
    <header className="flex items-center justify-between px-6 py-4 border-b border-border bg-card/50">
      <div className="flex items-center gap-4">
        <h2 className="font-display text-2xl font-semibold">{title}</h2>
        <span className="flex items-center gap-1.5 text-sm text-muted-foreground">
          <Calendar className="w-4 h-4" />
          {format(today, 'EEEE, MMMM d, yyyy')}
        </span>

        {/* Mood button */}
        <div className="relative" ref={moodPickerRef}>
          <button
            onClick={() => setShowMoodPicker(!showMoodPicker)}
            title="Set mood"
            className={cn(
              'p-2 rounded-lg transition-colors flex items-center gap-1',
              'bg-secondary/50 hover:bg-secondary text-muted-foreground hover:text-foreground'
            )}
          >
            {currentMood ? getMoodEmoji(currentMood) : <Smile className="w-4 h-4" />}
          </button>
          {showMoodPicker && (
            <div className="absolute top-full left-0 mt-1 bg-card border border-border rounded-lg shadow-lg z-50 p-2 flex gap-2">
              {MOOD_OPTIONS.map((option) => (
                <button
                  key={option.value}
                  onClick={() => handleMoodSelect(option.value)}
                  className="p-2 hover:bg-secondary/50 rounded transition-colors text-lg"
                >
                  {option.emoji}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Weather button */}
        <div className="relative" ref={weatherPickerRef}>
          <button
            onClick={() => setShowWeatherPicker(!showWeatherPicker)}
            title="Set weather"
            className={cn(
              'p-2 rounded-lg transition-colors flex items-center gap-1',
              'bg-secondary/50 hover:bg-secondary text-muted-foreground hover:text-foreground'
            )}
          >
            {currentWeather ? getWeatherEmoji(currentWeather) : <Cloud className="w-4 h-4" />}
          </button>
          {showWeatherPicker && (
            <div className="absolute top-full left-0 mt-1 bg-card border border-border rounded-lg shadow-lg z-50 p-2 flex gap-2">
              {WEATHER_OPTIONS.map((option) => (
                <button
                  key={option.value}
                  onClick={() => handleWeatherSelect(option.value)}
                  className="p-2 hover:bg-secondary/50 rounded transition-colors text-lg"
                >
                  {option.emoji}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Location button */}
        <div className="relative" ref={locationPickerRef}>
          <button
            onClick={() => setShowLocationPicker(!showLocationPicker)}
            title="Set location"
            className={cn(
              'p-2 rounded-lg transition-colors flex items-center gap-1',
              'bg-secondary/50 hover:bg-secondary text-muted-foreground hover:text-foreground'
            )}
          >
            {currentLocation ? (
              getLocationEmoji(currentLocation) || <span className="text-sm">{currentLocation}</span>
            ) : (
              <MapPin className="w-4 h-4" />
            )}
          </button>
          {showLocationPicker && (
            <div className="absolute top-full left-0 mt-1 bg-card border border-border rounded-lg shadow-lg z-50 p-2 w-48">
              {/* Quick location options */}
              <div className="flex gap-2 mb-2 pb-2 border-b border-border">
                {LOCATION_OPTIONS.map((option) => (
                  <button
                    key={option.value}
                    onClick={() => handleLocationSelect(option.value)}
                    className="p-2 hover:bg-secondary/50 rounded transition-colors text-lg"
                  >
                    {option.emoji}
                  </button>
                ))}
              </div>
              <input
                type="text"
                value={locationInput}
                onChange={(e) => setLocationInput(e.target.value)}
                onKeyDown={handleLocationKeyDown}
                placeholder="Enter location..."
                className="w-full px-2 py-1 text-sm rounded border border-border bg-background focus:outline-none focus:ring-1 focus:ring-primary/50 mb-2"
                autoFocus
              />
              {locationHistory.length > 0 && (
                <div className="space-y-1">
                  {locationHistory.map((loc) => (
                    <button
                      key={loc}
                      onClick={() => handleLocationSelect(loc)}
                      className="w-full text-left px-2 py-1 text-sm hover:bg-secondary/50 rounded transition-colors"
                    >
                      {loc}
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Capture button */}
      {onCapture && (
        <button
          onClick={onCapture}
          title="Capture entries"
          className={cn(
            'p-2 rounded-lg transition-colors',
            'bg-secondary/50 hover:bg-secondary text-muted-foreground hover:text-foreground'
          )}
        >
          <FileEdit className="w-4 h-4" />
        </button>
      )}
    </header>
  );
}
