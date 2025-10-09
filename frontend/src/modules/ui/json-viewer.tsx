'use client';

import { useState } from 'react';
import { Button } from '@modules/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './card';
import { cn } from '@modules/utils';

export function JsonViewer({ data, label = 'JSON' }: { data: unknown; label?: string }) {
	const [open, setOpen] = useState(true);

	return (
		<Card className={cn('flex-1 min-h-0 overflow-hidden', open ? 'h-[85dvh]' : '')}>
			<CardHeader className="justify-between flex-row flex">
				<CardTitle>{label}</CardTitle>
				<Button size="sm" variant="outline" onClick={() => setOpen(!open)}>
					{open ? 'Hide' : 'Show'}
				</Button>
			</CardHeader>

			<CardContent className="overflow-auto w-full">
				{!!open && (
					<div className="overflow-hidden flex h-full flex-col">
						<div className="flex-1 min-h-0 overflow-auto">{<pre className="text-xs p-3">{JSON.stringify(data, null, 4)}</pre>}</div>
					</div>
				)}
			</CardContent>
		</Card>
	);
}
