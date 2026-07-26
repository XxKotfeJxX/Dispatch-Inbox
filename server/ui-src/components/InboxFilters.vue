<script>
import { mailbox } from "../stores/mailbox";

const PLATFORM_PREFIX = "platform-";
const PRIORITIES = ["critical", "high", "normal", "low"];

export default {
	data() {
		return {
			mailbox,
			priorityOptions: PRIORITIES,
			open: false,
			platforms: [],
			priorities: [],
			state: "all",
			attachments: false,
			dateFrom: "",
			dateTo: "",
		};
	},

	computed: {
		availablePlatforms() {
			return mailbox.tags
				.filter((tag) => tag.toLowerCase().startsWith(PLATFORM_PREFIX))
				.map((tag) => tag.slice(PLATFORM_PREFIX.length))
				.sort();
		},

		activeCount() {
			return (
				this.platforms.length +
				this.priorities.length +
				(this.state === "all" ? 0 : 1) +
				(this.attachments ? 1 : 0) +
				(this.dateFrom ? 1 : 0) +
				(this.dateTo ? 1 : 0)
			);
		},
	},

	watch: {
		$route() {
			this.fromRoute();
		},
	},

	mounted() {
		this.fromRoute();
	},

	methods: {
		fromRoute() {
			const query = this.$route.query;
			this.platforms = typeof query.platforms === "string" ? query.platforms.split(",").filter(Boolean) : [];
			this.priorities = typeof query.priorities === "string" ? query.priorities.split(",").filter(Boolean) : [];
			this.state = ["read", "unread"].includes(query.state) ? query.state : "all";
			this.attachments = query.attachments === "1";
			this.dateFrom = typeof query.from === "string" ? query.from : "";
			this.dateTo = typeof query.to === "string" ? query.to : "";
		},

		buildSearch() {
			const terms = [];
			const text = typeof this.$route.query.text === "string" ? this.$route.query.text.trim() : "";
			if (text) terms.push(text);
			if (this.platforms.length) {
				terms.push(`tag-any:${this.platforms.map((v) => PLATFORM_PREFIX + v).join(",")}`);
			}
			if (this.priorities.length) {
				terms.push(`tag-any:${this.priorities.map((v) => "priority-" + v).join(",")}`);
			}
			if (this.state !== "all") terms.push(`is:${this.state}`);
			if (this.attachments) terms.push("has:attachment");
			if (this.dateFrom) terms.push(`after:${this.dateFrom}`);
			if (this.dateTo) terms.push(`before:${this.dateTo}`);
			return terms.join(" ");
		},

		apply() {
			const q = this.buildSearch();
			const routeQuery = {
				...(this.platforms.length ? { platforms: this.platforms.join(",") } : {}),
				...(this.priorities.length ? { priorities: this.priorities.join(",") } : {}),
				...(this.state !== "all" ? { state: this.state } : {}),
				...(this.attachments ? { attachments: "1" } : {}),
				...(this.dateFrom ? { from: this.dateFrom } : {}),
				...(this.dateTo ? { to: this.dateTo } : {}),
				...(q ? { q } : {}),
			};
			this.$router.push(q ? { path: "/search", query: routeQuery } : { path: "/" });
		},

		reset() {
			this.platforms = [];
			this.priorities = [];
			this.state = "all";
			this.attachments = false;
			this.dateFrom = "";
			this.dateTo = "";
			this.$router.push("/");
		},
	},
};
</script>

<template>
	<section class="inbox-tools mx-2 mt-3">
		<div class="d-flex flex-wrap align-items-center gap-2">
			<button class="btn btn-outline-primary" type="button" @click="open = !open">
				<i class="bi bi-funnel me-2"></i>
				Filters
				<span v-if="activeCount" class="badge rounded-pill text-bg-primary ms-2">{{ activeCount }}</span>
			</button>
			<button
				class="btn"
				:class="mailbox.summaryMode ? 'btn-primary' : 'btn-outline-secondary'"
				type="button"
				@click="mailbox.summaryMode = !mailbox.summaryMode"
			>
				<i class="bi bi-card-text me-2"></i>Summary
			</button>
			<select v-model="mailbox.sortOrder" class="form-select inbox-sort" aria-label="Sort messages">
				<option value="newest">Newest first</option>
				<option value="oldest">Oldest first</option>
				<option value="priority">Priority</option>
				<option value="size">Largest first</option>
			</select>
			<span class="ms-auto text-muted small">{{ mailbox.count }} matching messages</span>
		</div>

		<div v-if="open" class="filter-panel mt-3">
			<div class="filter-group">
				<div class="filter-title">Platform</div>
				<label v-for="platform in availablePlatforms" :key="platform" class="form-check">
					<input v-model="platforms" class="form-check-input" type="checkbox" :value="platform" />
					<span class="form-check-label text-capitalize">{{ platform }}</span>
				</label>
				<span v-if="!availablePlatforms.length" class="small text-muted"
					>Platforms appear after first delivery.</span
				>
			</div>
			<div class="filter-group">
				<div class="filter-title">Priority</div>
				<label v-for="priority in priorityOptions" :key="priority" class="form-check">
					<input v-model="priorities" class="form-check-input" type="checkbox" :value="priority" />
					<span class="form-check-label text-capitalize">{{ priority }}</span>
				</label>
			</div>
			<div class="filter-group">
				<label class="form-label filter-title" for="message-state">Status</label>
				<select id="message-state" v-model="state" class="form-select">
					<option value="all">All</option>
					<option value="unread">Unread</option>
					<option value="read">Read</option>
				</select>
				<label class="form-check mt-3">
					<input v-model="attachments" class="form-check-input" type="checkbox" />
					<span class="form-check-label">Has attachments</span>
				</label>
			</div>
			<div class="filter-group">
				<div class="filter-title">Received</div>
				<label class="form-label small" for="date-from">From</label>
				<input id="date-from" v-model="dateFrom" class="form-control mb-2" type="date" />
				<label class="form-label small" for="date-to">To</label>
				<input id="date-to" v-model="dateTo" class="form-control" type="date" />
			</div>
			<div class="filter-actions">
				<button class="btn btn-primary" type="button" @click="apply">Apply filters</button>
				<button class="btn btn-link text-muted" type="button" @click="reset">Clear all</button>
			</div>
		</div>
	</section>
</template>
