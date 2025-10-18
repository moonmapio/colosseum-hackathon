'use client';

import * as React from 'react';
import { Textarea } from '@modules/ui/textarea';
import { FieldError } from './FieldError';
import { cn } from '@modules/utils';

type Props = React.ComponentProps<typeof Textarea> & {
	labelClassName?: string;
	label: string;
	error?: string;
	children?: React.ReactNode;
	showCounts?: boolean;
	maxWords?: number;
	maxLetters?: number;
	onCounts?: (words: number, letters: number) => void;
};

function countWords(v: string) {
	const m = v.trim().match(/[\p{L}\p{N}]+/gu);
	return m ? m.length : 0;
}

function countLetters(v: string) {
	const m = v.match(/\p{L}/gu);
	return m ? m.length : 0;
}

export function LabeledTextarea({ label, error, children, showCounts = true, maxWords, maxLetters, onCounts, ...rest }: Props) {
	const { className, onChange, defaultValue, value, labelClassName, ...inputProps } = rest;
	const initial = typeof value === 'string' ? value : typeof defaultValue === 'string' ? defaultValue : '';
	const [words, setWords] = React.useState(() => countWords(String(initial)));
	const [letters, setLetters] = React.useState(() => countLetters(String(initial)));

	React.useEffect(() => {
		if (typeof value === 'string') {
			setWords(countWords(value));
			setLetters(countLetters(value));
		}
	}, [value]);

	const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
		const v = e.target.value;
		const w = countWords(v);
		const l = countLetters(v);
		setWords(w);
		setLetters(l);
		onCounts?.(w, l);
		onChange?.(e);
	};

	const overWords = typeof maxWords === 'number' && words > maxWords;
	const overLetters = typeof maxLetters === 'number' && letters > maxLetters;
	const over = overWords || overLetters;

	return (
		<div className=" h-36">
			<label className={cn('text-sm flex-row flex font-medium items-center', labelClassName)}>
				<div className="mr-1">{children}</div>
				{label}
			</label>
			<div className="relative">
				<Textarea
					{...inputProps}
					defaultValue={defaultValue}
					value={value}
					onChange={handleChange}
					className={cn('bg-background !text-xs mt-1 h-full', className, over && 'ring-1 ring-destructive')}
				/>
				{showCounts && (
					<div className="pointer-events-none absolute -bottom-6 right-0 text-[10px] leading-none select-none flex gap-2 bg-secondary text-primary p-1 rounded-md">
						{words > 0 && (
							<span className={cn(overWords ? 'text-destructive' : 'text-muted-foreground')}>
								{words}
								{typeof maxWords === 'number' ? ` / ${maxWords}` : ''} {words > 1 ? 'words' : 'word'}
							</span>
						)}
						{letters > 0 && (
							<span className={cn(overLetters ? 'text-destructive' : 'text-muted-foreground')}>
								{letters}
								{typeof maxLetters === 'number' ? ` / ${maxLetters}` : ''} {letters > 1 ? 'characters' : 'character'}
							</span>
						)}
					</div>
				)}
			</div>
			<FieldError msg={error} />
		</div>
	);
}
