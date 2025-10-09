import { cn } from '@modules/utils';
import z from 'zod';

import { SiDiscord, SiTelegram, SiX } from '@icons-pack/react-simple-icons';
import { ProjectChain } from '@modules/dtos/common';
import { CreateProjectSchema } from '@modules/dtos/project';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@modules/ui/accordion';
import { CaseSensitive, Coins, Fingerprint, Globe2, QrCode, ScanText, Users2 } from 'lucide-react';
import { useFormContext, useWatch } from 'react-hook-form';
import { LabeledInput } from './LabeledInput';
import { LabeledSelect } from './LabeledSelect';
import { LabeledTextarea } from './LabeledTextarea';
import { LogoPicker } from './LogoPicker';
import { LaunchDateField } from './LaunchDateField';

export type CreateProjectInput = z.infer<typeof CreateProjectSchema>;

type ComputeFunction = (watchedValue: string | null | undefined) => boolean;

export function DialogBody() {
	const form = useFormContext<CreateProjectInput>();
	const control = form.control;

	const compute: ComputeFunction = (watchedValue) => Boolean(watchedValue ? watchedValue.length > 0 : 0);

	const twitterActive = useWatch<CreateProjectInput>({ name: 'twitter', compute, control });
	const discordActive = useWatch<CreateProjectInput>({ name: 'discord', compute, control });
	const telegramActive = useWatch<CreateProjectInput>({ name: 'telegram', compute, control });
	const websiteActive = useWatch<CreateProjectInput>({ name: 'website', compute, control });

	return (
		<div className="!space-y-4 w-full transition-all">
			<div className="grid md:grid-cols-2 gap-4 min-w-0">
				<div className="rounded-xl bg-muted p-4 min-w-0 flex flex-col justify-between h-[440px]">
					<div>
						<div className="!text-md leading-none font-semibold flex flex-row items-center w-full mb-2">
							<span className="mr-2">
								<QrCode size={20} />
							</span>
							Basic Coin Data
						</div>
						<p className="text-xs text-muted-foreground">Click the circle to choose an image from your computer.</p>
					</div>

					<div>
						<div className="grid md:grid-cols-2 gap-4 min-w-0">
							<div className="min-w-0 max-w-full">
								<LogoPicker className="items-center justify-end h-full flex flex-col" />
							</div>
							<div className="flex h-full flex-col gap-y-2 justify-end min-w-0">
								<LabeledInput
									label="Name"
									placeholder="Example Coin"
									{...form.register('name')}
									error={form.formState.errors.name?.message}
								>
									<CaseSensitive className="size-3.5" />
								</LabeledInput>
								<LabeledInput label="Ticker" placeholder="EXMPL" {...form.register('symbol')} error={form.formState.errors.symbol?.message}>
									<Coins className="size-3.5" />
								</LabeledInput>
								<LabeledSelect
									label="Chain"
									defaultValue="solana"
									disabled={true}
									{...form.register('chain')}
									error={form.formState.errors.chain?.message}
								>
									{Object.values(ProjectChain).map((c) => (
										<option key={c} value={c} className="!capitalize">
											{c.toLocaleLowerCase()}
										</option>
									))}
								</LabeledSelect>
							</div>
						</div>
						<div className="mt-3">
							<LaunchDateField />
						</div>
						<div className="mt-3">
							<LabeledInput
								label="Contract Address"
								placeholder="moonXxaaJxmmoonMap"
								{...form.register('contractAddress')}
								error={form.formState.errors.contractAddress?.message}
							>
								<Fingerprint className="size-3.5" />
							</LabeledInput>
						</div>
					</div>
				</div>

				<div className="flex flex-col gap-y-2 min-w-0">
					<div className="rounded-xl bg-muted p-4 h-[215px]">
						<LabeledTextarea
							rows={3}
							label="Narrative (optional)"
							placeholder="Why this token matters…"
							{...form.register('narrative')}
							error={form.formState.errors.narrative?.message}
							labelClassName="!text-md mb-2"
							className="max-h-36"
						>
							<ScanText size={20} />
						</LabeledTextarea>
					</div>

					<Accordion type="single" collapsible>
						<AccordionItem value="social">
							<div className="rounded-xl bg-muted p-4">
								<AccordionTrigger className="p-0 h-5 hover:cursor-pointer w-full">
									<div className="!text-md leading-none font-semibold flex flex-row items-center w-full">
										<span className="mr-2">
											<Users2 size={20} />
										</span>
										Community Channels
										<span className="flex flex-row gap-x-2 items-center ml-auto">
											<SiX className={cn('size-3.5 text-muted-foreground/50', twitterActive ? 'text-primary' : '')} />
											<SiTelegram className={cn('size-4 text-muted-foreground/50', telegramActive ? 'text-primary' : '')} />
											<SiDiscord className={cn('size-4 text-muted-foreground/50', discordActive ? 'text-primary' : '')} />
											<Globe2 className={cn('size-4 text-muted-foreground/50', websiteActive ? 'text-primary' : '')} />
										</span>
									</div>
								</AccordionTrigger>
								<AccordionContent className="p-1 mb-1 overflow-hidden">
									<div className="mt-2 grid md:grid-cols-2 gap-2 max-h-56 overflow-auto p-1 min-w-0">
										<LabeledInput
											label="Twitter"
											placeholder="https://x.com/user"
											{...form.register('twitter')}
											error={form.formState.errors.twitter?.message}
										>
											<SiX className="size-3" />
										</LabeledInput>
										<LabeledInput
											label="Telegram"
											placeholder="https://t.me/…"
											{...form.register('telegram')}
											error={form.formState.errors.telegram?.message}
										>
											<SiTelegram className="size-3" />
										</LabeledInput>
										<LabeledInput
											label="Discord"
											placeholder="https://discord.gg/…"
											{...form.register('discord')}
											error={form.formState.errors.discord?.message}
										>
											<SiDiscord className="size-3" />
										</LabeledInput>
										<LabeledInput
											label="Website"
											placeholder="https:/moonmap.io"
											{...form.register('website')}
											error={form.formState.errors.website?.message}
										>
											<Globe2 className="size-3.5" />
										</LabeledInput>
									</div>
								</AccordionContent>
							</div>
						</AccordionItem>
					</Accordion>
				</div>
			</div>
		</div>
	);
}
