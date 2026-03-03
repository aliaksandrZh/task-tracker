import React, { useState, useMemo } from 'react';
import { Box, Text, useInput } from 'ink';
import { loadTasks } from '../store.js';
import { groupByDate, filterCurrentWeek } from '../utils.js';
import TaskTable from './TaskTable.jsx';

export default function ViewSummary({ onDone }) {
  const [mode, setMode] = useState('daily');
  const tasks = useMemo(() => loadTasks(), []);
  const dailyGroups = useMemo(() => groupByDate(tasks), [tasks]);
  const weeklyTasks = useMemo(() => filterCurrentWeek(tasks), [tasks]);

  useInput((ch, key) => {
    if (key.escape) {
      onDone();
      return;
    }
    if (ch === 'd') setMode('daily');
    if (ch === 'w') setMode('weekly');
  });

  if (tasks.length === 0) {
    return (
      <Box flexDirection="column">
        <Text>No tasks found. Press Escape to go back.</Text>
      </Box>
    );
  }

  if (mode === 'weekly') {
    return (
      <Box flexDirection="column">
        <Text bold>Weekly Summary</Text>
        <Text dimColor>d = daily | w = weekly | Escape = back</Text>

        <Box marginTop={1} flexDirection="column">
          <Text bold color="yellow">
            {weeklyTasks.label} — {weeklyTasks.total.toFixed(1)}h total ({weeklyTasks.tasks.length} tasks)
          </Text>
          {weeklyTasks.tasks.length > 0
            ? <TaskTable tasks={weeklyTasks.tasks} />
            : <Text color="gray">No tasks this week.</Text>
          }
        </Box>
      </Box>
    );
  }

  return (
    <Box flexDirection="column">
      <Text bold>Daily Summary</Text>
      <Text dimColor>d = daily | w = weekly | Escape = back</Text>

      {dailyGroups.map(group => (
        <Box key={group.key} marginTop={1} flexDirection="column">
          <Text bold color="yellow">
            {group.key} — {group.total.toFixed(1)}h total ({group.tasks.length} tasks)
          </Text>
          <TaskTable tasks={group.tasks} />
        </Box>
      ))}
    </Box>
  );
}
