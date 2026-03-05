import React, { useState, useMemo } from 'react';
import { Box, Text, useInput } from 'ink';
import { loadTasks } from '../store.js';
import { groupByDate, filterWeekByOffset, parseDate, getWeekBounds, sortTasks } from '../utils.js';
import { loadPrefs, savePrefs } from '../prefs.js';
import TaskTable from './TaskTable.jsx';

function getMinWeekOffset(tasks) {
  let earliest = null;
  for (const t of tasks) {
    const d = parseDate(t.date);
    if (d && (earliest === null || d < earliest)) earliest = d;
  }
  if (!earliest) return 0;
  const now = new Date();
  const currentMonday = getWeekBounds(now).monday;
  const earliestMonday = getWeekBounds(earliest).monday;
  const diffMs = earliestMonday - currentMonday;
  return Math.floor(diffMs / (7 * 24 * 60 * 60 * 1000));
}

export default function ViewSummary({ onDone }) {
  const SORT_COLUMNS = [null, 'date', 'type', 'number', 'name', 'timeSpent'];
  const prefs = useMemo(() => loadPrefs(), []);
  const [mode, setMode] = useState('daily');
  const [dailyIdx, setDailyIdx] = useState(0);
  const [weekOffset, setWeekOffset] = useState(0);
  const [sortBy, setSortBy] = useState(prefs.sortBy || null);
  const [sortDir, setSortDir] = useState(prefs.sortDir || 'asc');
  const tasks = useMemo(() => loadTasks(), []);
  const dailyGroups = useMemo(() => groupByDate(tasks), [tasks]);
  const minWeekOffset = useMemo(() => getMinWeekOffset(tasks), [tasks]);

  const clearScreen = () => process.stdout.write('\x1B[2J\x1B[H');

  useInput((ch, key) => {
    if (key.escape) {
      onDone();
      return;
    }
    if (ch === 'd' && mode !== 'daily') { clearScreen(); setMode('daily'); }
    if (ch === 'w' && mode !== 'weekly') { clearScreen(); setMode('weekly'); }
    if (ch === 's') {
      clearScreen();
      setSortBy(prev => {
        const idx = SORT_COLUMNS.indexOf(prev);
        const next = SORT_COLUMNS[(idx + 1) % SORT_COLUMNS.length];
        savePrefs({ sortBy: next });
        return next;
      });
    }
    if (ch === 'S') {
      clearScreen();
      setSortDir(prev => {
        const next = prev === 'asc' ? 'desc' : 'asc';
        savePrefs({ sortDir: next });
        return next;
      });
    }

    if (key.leftArrow) {
      if (mode === 'daily' && dailyIdx < dailyGroups.length - 1) {
        clearScreen();
        setDailyIdx(i => i + 1);
      } else if (mode === 'weekly' && weekOffset > minWeekOffset) {
        clearScreen();
        setWeekOffset(o => o - 1);
      }
    }
    if (key.rightArrow) {
      if (mode === 'daily' && dailyIdx > 0) {
        clearScreen();
        setDailyIdx(i => i - 1);
      } else if (mode === 'weekly' && weekOffset < 0) {
        clearScreen();
        setWeekOffset(o => o + 1);
      }
    }
  });

  if (tasks.length === 0) {
    return (
      <Box flexDirection="column">
        <Text>No tasks found. Press Escape to go back.</Text>
      </Box>
    );
  }

  const sortHint = sortBy ? `sort:${sortBy} ${sortDir}` : 'off';
  const hint = `← prev | → next | d=daily | w=weekly | s=sort(${sortHint}) | S=flip | Escape=back`;

  if (mode === 'weekly') {
    const weeklyTasks = filterWeekByOffset(tasks, weekOffset);
    return (
      <Box flexDirection="column">
        <Text bold>Weekly Summary</Text>
        <Text dimColor>{hint}</Text>

        <Box marginTop={1} flexDirection="column">
          <Text bold color="yellow">
            {weeklyTasks.label} — {weeklyTasks.total.toFixed(1)}h total ({weeklyTasks.tasks.length} tasks)
          </Text>
          {weeklyTasks.tasks.length > 0
            ? <TaskTable tasks={sortTasks(weeklyTasks.tasks, sortBy, sortDir)} sortBy={sortBy} sortDir={sortDir} />
            : <Text color="gray">No tasks this week.</Text>
          }
        </Box>
      </Box>
    );
  }

  const group = dailyGroups[dailyIdx];

  return (
    <Box flexDirection="column">
      <Text bold>Daily Summary — {group ? group.key : ''}</Text>
      <Text dimColor>{hint}</Text>

      {group && (
        <Box marginTop={1} flexDirection="column">
          <Text bold color="yellow">
            {group.key} — {group.total.toFixed(1)}h total ({group.tasks.length} tasks)
          </Text>
          <TaskTable tasks={sortTasks(group.tasks, sortBy, sortDir)} sortBy={sortBy} sortDir={sortDir} />
        </Box>
      )}
    </Box>
  );
}
