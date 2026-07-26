<script>
import CommonMixins from "./mixins/CommonMixins";
import Favicon from "./components/AppFavicon.vue";
import AppBadge from "./components/AppBadge.vue";
import Notifications from "./components/AppNotifications.vue";
import EditTags from "./components/EditTags.vue";
import { mailbox } from "./stores/mailbox";

export default {
	components: {
		Favicon,
		AppBadge,
		Notifications,
		EditTags,
	},

	mixins: [CommonMixins],

	watch: {
		$route() {
			// hide mobile menu on URL change
			this.hideNav();
		},
		"mailbox.autoDeleteReadHours"() {
			this.runReadRetention();
		},
	},

	beforeMount() {
		// load global config
		this.get(this.resolve("/api/v1/webui"), false, (response) => {
			mailbox.uiConfig = response.data;

			if (mailbox.uiConfig.Label) {
				document.title = document.title + " - " + mailbox.uiConfig.Label;
			} else {
				document.title = document.title + " - " + location.hostname;
			}
		});
		this.retentionTimer = window.setInterval(() => this.runReadRetention(), 5 * 60 * 1000);
		window.setTimeout(() => this.runReadRetention(), 2500);
	},

	beforeUnmount() {
		window.clearInterval(this.retentionTimer);
	},

	methods: {
		runReadRetention() {
			const hours = Number(mailbox.autoDeleteReadHours);
			if (!hours || hours < 1) return;

			this.get(
				this.resolve("/api/v1/search") + "?query=" + encodeURIComponent("is:read"),
				{ limit: 1000 },
				(response) => {
					const cutoff = Date.now() - hours * 60 * 60 * 1000;
					const IDs = response.data.messages
						.filter((message) => new Date(message.Created).getTime() < cutoff)
						.map((message) => message.ID);
					if (!IDs.length) return;
					this.delete(this.resolve("/api/v1/messages"), { IDs }, () => {
						mailbox.refresh = true;
					});
				},
			);
		},
	},
};
</script>

<template>
	<RouterView />
	<Favicon />
	<AppBadge />
	<Notifications />
	<EditTags />
</template>
