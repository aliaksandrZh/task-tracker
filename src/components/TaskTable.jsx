import React from 'react';
import { Box, Text, useStdout } from 'ink';
import { pad, FIXED, MIN_NAME, MIN_COMMENTS, GAP } from '../format.js';

const TYPE_COLORS = {
  bug: 'red',
  task: 'yellow',
};

function getTypeColor(type) {
  return TYPE_COLORS[(type || '').toLowerCase()] || 'white';
}

export default function TaskTable({ tasks, showIndex = false, sortBy = null, sortDir = 'asc' }) {
  const { stdout } = useStdout();
  const width = stdout?.columns || 80;

  if (!tasks || tasks.length === 0) {
    return <Text color="gray">No tasks to display.</Text>;
  }

  const numCols = showIndex ? 7 : 6;
  const numGaps = numCols - 1;
  const fixedWidth = FIXED.date + FIXED.type + FIXED.number + FIXED.timeSpent + (showIndex ? 4 : 0) + numGaps * GAP.length;
  const remaining = Math.max(MIN_NAME + MIN_COMMENTS, width - fixedWidth);
  const nameWidth = Math.floor(remaining * 0.65);
  const commentsWidth = remaining - nameWidth;

  const join = (...parts) => parts.join(GAP);

  const indicator = sortDir === 'asc' ? '▲' : '▼';
  const h = (label, col, w) => pad(sortBy === col ? `${label}${indicator}` : label, w);

  const headerText = join(
    ...(showIndex ? [pad('#', 4)] : []),
    h('Date', 'date', FIXED.date),
    h('Type', 'type', FIXED.type),
    h('Number', 'number', FIXED.number),
    h('Name', 'name', nameWidth),
    h('Time', 'timeSpent', FIXED.timeSpent),
    pad('Comments', commentsWidth),
  );

  return (
    <Box flexDirection="column" width={width}>
      <Text bold color="cyan">{headerText}</Text>
      {tasks.map((t, i) => {
        const beforeType = join(
          ...(showIndex ? [pad(String(i + 1), 4)] : []),
          pad(t.date, FIXED.date),
        );
        const afterType = join(
          pad(t.number, FIXED.number),
          pad(t.name, nameWidth),
          pad(t.timeSpent, FIXED.timeSpent),
          pad(t.comments, commentsWidth),
        );
        return (
          <Text key={i}>
            <Text>{beforeType}{GAP}</Text>
            <Text color={getTypeColor(t.type)}>{pad(t.type, FIXED.type)}</Text>
            <Text>{GAP}{afterType}</Text>
          </Text>
        );
      })}
    </Box>
  );
}
