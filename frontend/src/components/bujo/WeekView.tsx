import { useEffect, useState } from 'react';
import { Entry } from '@/types/bujo';
import { getEntriesForDateRange } from '@/api/bujo';
import { filterWeekEntries } from '@/lib/weekView';
import { DayBox } from './DayBox';
import { WeekendBox } from './WeekendBox';
import { JournalSidebar } from './JournalSidebar';
import { addDays, format } from 'date-fns';

interface WeekViewProps {
  startDate: Date;
}

const DAYS_IN_WEEK = 7;

export function WeekView({ startDate }: WeekViewProps) {
  const [entries, setEntries] = useState<Entry[]>([]);
  const [selectedEntryId, setSelectedEntryId] = useState<number | null>(null);

  useEffect(() => {
    const endDate = addDays(startDate, DAYS_IN_WEEK - 1);
    const startStr = format(startDate, 'yyyy-MM-dd');
    const endStr = format(endDate, 'yyyy-MM-dd');

    getEntriesForDateRange(startStr, endStr).then(data => {
      setEntries(data);
    });
  }, [startDate]);

  const filteredEntries = filterWeekEntries(entries);

  const getEntriesForDay = (dayOffset: number): Entry[] => {
    const day = addDays(startDate, dayOffset);
    const dayStr = format(day, 'yyyy-MM-dd');
    return filteredEntries.filter(e => e.loggedDate === dayStr);
  };

  const monday = getEntriesForDay(0);
  const tuesday = getEntriesForDay(1);
  const wednesday = getEntriesForDay(2);
  const thursday = getEntriesForDay(3);
  const friday = getEntriesForDay(4);
  const saturday = getEntriesForDay(5);
  const sunday = getEntriesForDay(6);

  const selectedEntry = selectedEntryId
    ? filteredEntries.find(e => e.id === selectedEntryId) || null
    : null;

  const handleSelectEntry = (entry: Entry) => {
    setSelectedEntryId(entry.id);
  };

  return (
    <div className="flex h-full gap-4">
      <div className="flex-1">
        <div className="grid grid-cols-2 gap-4">
          <DayBox
            date={startDate}
            entries={monday}
            selectedEntryId={selectedEntryId || undefined}
            onEntrySelect={setSelectedEntryId}
          />
          <DayBox
            date={addDays(startDate, 1)}
            entries={tuesday}
            selectedEntryId={selectedEntryId || undefined}
            onEntrySelect={setSelectedEntryId}
          />
          <DayBox
            date={addDays(startDate, 2)}
            entries={wednesday}
            selectedEntryId={selectedEntryId || undefined}
            onEntrySelect={setSelectedEntryId}
          />
          <DayBox
            date={addDays(startDate, 3)}
            entries={thursday}
            selectedEntryId={selectedEntryId || undefined}
            onEntrySelect={setSelectedEntryId}
          />
          <DayBox
            date={addDays(startDate, 4)}
            entries={friday}
            selectedEntryId={selectedEntryId || undefined}
            onEntrySelect={setSelectedEntryId}
          />
          <WeekendBox
            startDay={parseInt(format(addDays(startDate, 5), 'd'))}
            saturdayEntries={saturday}
            sundayEntries={sunday}
            selectedEntry={selectedEntry || undefined}
            onSelectEntry={handleSelectEntry}
          />
        </div>
      </div>

      {selectedEntry && (
        <div className="w-80 border-l pl-4">
          <JournalSidebar
            selectedEntry={selectedEntry}
            contextTree={entries}
          />
        </div>
      )}
    </div>
  );
}
