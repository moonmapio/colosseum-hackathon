'use client';
import { Calendar } from '@modules/ui/calendar';
import { SidebarGroup, SidebarGroupContent } from '@modules/ui/sidebar';
import { useState } from 'react';

export function DatePicker() {
	const [date, setDate] = useState<Date | undefined>(new Date(2025, 5, 12));

	return (
		<SidebarGroup className="px-0">
			<SidebarGroupContent>
				<Calendar
					mode="single"
					defaultMonth={date}
					selected={date}
					onSelect={setDate}
					disabled={{
						before: new Date(2025, 5, 12),
					}}
					className="[&_[role=gridcell].bg-accent]:bg-sidebar-primary [&_[role=gridcell].bg-accent]:text-sidebar-primary-foreground [&_[role=gridcell]]:w-[33px]"
				/>
			</SidebarGroupContent>
		</SidebarGroup>
	);
}
