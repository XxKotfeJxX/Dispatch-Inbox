<script>
import { mailbox } from "../stores/mailbox";
import CommonMixins from "../mixins/CommonMixins";
import dayjs from "dayjs";
import { pagination } from "../stores/pagination";

export default {
	mixins: [CommonMixins],

	props: {
		// use different name to `loading` as that is already in use in CommonMixins
		loadingMessages: {
			type: Number,
			default: 0,
		},
	},

	data() {
		return {
			mailbox,
			pagination,
		};
	},

	computed: {
		sortedMessages() {
			const messages = [...mailbox.messages];
			const priority = { critical: 0, high: 1, normal: 2, low: 3 };
			const getPriority = (message) => {
				const tag = message.Tags.find((value) => value.toLowerCase().startsWith("priority-"));
				return tag ? tag.slice("priority-".length).toLowerCase() : "normal";
			};

			if (mailbox.sortOrder === "oldest") {
				return messages.sort((a, b) => new Date(a.Created) - new Date(b.Created));
			}
			if (mailbox.sortOrder === "size") {
				return messages.sort((a, b) => b.Size - a.Size);
			}
			if (mailbox.sortOrder === "priority") {
				return messages.sort((a, b) => priority[getPriority(a)] - priority[getPriority(b)]);
			}
			return messages.sort((a, b) => new Date(b.Created) - new Date(a.Created));
		},
	},

	created() {
		const relativeTime = require("dayjs/plugin/relativeTime");
		dayjs.extend(relativeTime);
	},

	mounted() {
		this.refreshUI();
	},

	methods: {
		refreshUI() {
			window.setTimeout(() => {
				this.$forceUpdate();
				this.refreshUI();
			}, 30000);
		},

		getRelativeCreated(message) {
			const d = new Date(message.Created);
			return dayjs(d).fromNow();
		},

		getPrimaryEmailTo(message) {
			if (message.To && message.To.length > 0) {
				return message.To[0].Address;
			}

			return "[ Undisclosed recipients ]";
		},

		getPlatform(message) {
			const tag = message.Tags.find((value) => value.toLowerCase().startsWith("platform-"));
			return tag ? tag.slice("platform-".length) : "Dispatch";
		},

		getPriority(message) {
			const tag = message.Tags.find((value) => value.toLowerCase().startsWith("priority-"));
			return tag ? tag.slice("priority-".length) : "normal";
		},

		isSelected(id) {
			return mailbox.selected.indexOf(id) !== -1;
		},

		toggleSelected(e, id) {
			e.preventDefault();

			if (this.isSelected(id)) {
				mailbox.selected = mailbox.selected.filter((ele) => {
					return ele !== id;
				});
			} else {
				mailbox.selected.push(id);
			}
		},

		selectRange(e, id) {
			e.preventDefault();

			let selecting = false;
			const lastSelected = mailbox.selected.length > 0 && mailbox.selected[mailbox.selected.length - 1];
			if (lastSelected === id) {
				mailbox.selected = mailbox.selected.filter((ele) => {
					return ele !== id;
				});
				return;
			}

			if (lastSelected === false) {
				mailbox.selected.push(id);
				return;
			}

			for (const d of mailbox.messages) {
				if (selecting) {
					if (!this.isSelected(d.ID)) {
						mailbox.selected.push(d.ID);
					}
					if (d.ID === lastSelected || d.ID === id) {
						// reached backwards select
						break;
					}
				} else if (d.ID === id || d.ID === lastSelected) {
					if (!this.isSelected(d.ID)) {
						mailbox.selected.push(d.ID);
					}
					selecting = true;
				}
			}
		},

		toTagUrl(t) {
			if (t.match(/ /)) {
				t = `"${t}"`;
			}
			const p = {
				q: "tag:" + t,
			};
			if (pagination.limit !== pagination.defaultLimit) {
				p.limit = pagination.limit.toString();
			}
			const params = new URLSearchParams(p);
			return "/search?" + params.toString();
		},
	},
};
</script>

<template>
	<template v-if="mailbox.messages && mailbox.messages.length">
		<div class="dispatch-message-list mx-2 my-3" :class="{ 'summary-grid': mailbox.summaryMode }">
			<RouterLink
				v-for="message in sortedMessages"
				:id="message.ID"
				:key="'message_' + message.ID"
				:to="'/view/' + message.ID"
				class="dispatch-message message"
				:class="[message.Read ? 'read' : '', isSelected(message.ID) ? ' selected' : '']"
				@click.meta="toggleSelected($event, message.ID)"
				@click.ctrl="toggleSelected($event, message.ID)"
				@click.shift="selectRange($event, message.ID)"
			>
				<div class="message-platform" :class="'platform-' + getPlatform(message).toLowerCase()">
					<i class="bi bi-app-indicator"></i>
					<span>{{ getPlatform(message) }}</span>
				</div>
				<div class="message-copy">
					<div class="d-flex align-items-center gap-2">
						<span
							class="priority-dot"
							:class="'priority-' + getPriority(message)"
							:title="getPriority(message)"
						></span>
						<div class="subject text-truncate text-spaces-nowrap">
							<b>{{ message.Subject !== "" ? message.Subject : "[ no subject ]" }}</b>
						</div>
					</div>
					<div
						v-if="message.Snippet !== ''"
						class="message-summary text-muted"
						:class="{ 'text-truncate': !mailbox.summaryMode }"
					>
						{{ message.Snippet }}
					</div>
					<div v-if="mailbox.summaryMode" class="message-origin text-muted small privacy">
						<span v-if="message.From">
							{{ message.From.Name || message.From.Address }}
						</span>
						<span v-if="message.To?.length"> → {{ getPrimaryEmailTo(message) }}</span>
					</div>
				</div>
				<div class="message-meta text-muted">
					<span v-if="message.Attachments" title="Has attachments"><i class="bi bi-paperclip"></i></span>
					<span>{{ getRelativeCreated(message) }}</span>
					<i class="bi bi-chevron-right"></i>
				</div>
			</RouterLink>
		</div>
	</template>
	<template v-else>
		<p class="text-center mt-5">
			<span v-if="loadingMessages > 0" class="text-muted"> Loading messages... </span>
			<template v-else-if="getSearch()"
				>No results for <code>{{ getSearch() }}</code></template
			>
			<template v-else>No messages in your mailbox</template>
		</p>
	</template>
</template>
