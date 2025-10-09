import { cn } from '@modules/utils';
import { TriangleAlert } from 'lucide-react';
import { useFormContext } from 'react-hook-form';
import { CreateProjectInput } from './DialogBody';

export type ErrorShape = { error?: string; message?: string };

export function ErrorBanner({ data }: { data: ErrorShape | null }) {
	if (!data?.error) return null;
	const isDup = data.error?.toUpperCase() === 'DUPLICATE_KEY';
	const base = 'rounded-lg px-4 py-3 text-sm';
	const cl = isDup ? 'bg-destructive/10 text-destructive' : 'bg-muted text-foreground';
	const msg = isDup ? 'A project with that wallet address already exists.' : data.message || 'Unexpected error.';
	return <div className={cn(base, cl, 'mb-3')}>{msg}</div>;
}

export function ErrorBannerFormErrors() {
	const form = useFormContext<CreateProjectInput>();
	const formErrors = form.formState.errors;

	if (!Object.values(formErrors).find((e) => !!e?.message?.length)) return;
	// if (Object.values(formErrors).length) return null;
	const base = 'rounded-lg px-4 py-3 text-sm';
	const cl = 'bg-primary text-primary-foreground';

	return (
		<div className={cn(base, cl, 'w-full mb-3')}>
			<div className="flex flex-row text-xs font-bold mb-2">
				<TriangleAlert className="size-4 mr-2" />
				<span>Please fix following errors first:</span>
			</div>
			<div className="flex flex-col text-xs ml-5">
				{Object.entries(formErrors).map(([fieldName, fieldError]) => {
					if (!fieldError?.message) return null;
					return (
						<li key={fieldName} className="break-all">
							<strong>{fieldName}:</strong> {fieldError.message}
						</li>
					);
				})}
			</div>
		</div>
	);
}
