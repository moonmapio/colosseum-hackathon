'use client';

import { Input } from '@modules/ui/input';
import { Tooltip, TooltipContent } from '@modules/ui/tooltip';
import { TooltipTrigger } from '@radix-ui/react-tooltip';
import * as React from 'react';
import { useController } from 'react-hook-form';
// import { FieldError } from './FieldError';

type Props = React.ComponentProps<typeof Input> & { label: string; error?: string; children?: React.ReactNode };
export function LabeledInput({ label, children, ...rest }: Props) {
	const value = useController({ name: rest.name ?? '' });
	return (
		<Tooltip>
			<div>
				<label className="text-sm flex-row flex font-medium items-center mb-2">
					<div className="mr-1">{children}</div>
					{label}
				</label>

				<TooltipTrigger asChild>
					<Input {...rest} className="bg-background !text-xs" />
					{/* <FieldError msg={error} /> */}
				</TooltipTrigger>
				{!!value.field.value && <TooltipContent className="w-[300px]">{value.field.value}</TooltipContent>}
			</div>
		</Tooltip>
	);
}
