import { cn } from '@modules/utils';
import { FieldError } from './FieldError';
import { ChevronDown, Link } from 'lucide-react';

type Props = React.SelectHTMLAttributes<HTMLSelectElement> & { label: string; error?: string };

export function LabeledSelect({ label, error, children, className, ...rest }: Props) {
	return (
		<div className="w-full">
			<label className="text-sm flex-row flex font-medium items-center mb-1">
				<div className="mr-1">
					<Link className="size-3.5" />
				</div>
				{label}
			</label>

			<div className="relative">
				<select
					className={cn(
						'w-full h-9 rounded-md border bg-background pl-3 pr-10 text-sm outline-none ring-offset-background appearance-none',
						'focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
						className,
					)}
					{...rest}
				>
					{children}
				</select>

				<ChevronDown className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
			</div>
			<FieldError msg={error} />
		</div>
	);
}
