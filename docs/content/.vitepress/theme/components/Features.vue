<template>
    <section class="flex flex-col gap-6">
        <h2 class="opacity-40 ms-8">WHAT'S INCLUDED</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div
                class="flex flex-col justify-between bg-[#424242] bg-opacity-40 border border-gray-600 py-8 px-6 rounded-xl gap-6 last:lg:col-span-2"
                v-for="feature in frontmatter.features"
            >
                <div class="flex flex-col gap-6">
                    <h3
                        class="font-semibold text-2xl"
                        :class="{
                            'text-left': feature.last,
                            'text-center': !feature.last,
                        }"
                    >
                        {{ feature.title }}
                    </h3>
                    <p class="font-light text-lg">
                        {{ feature.description }}
                    </p>
                </div>
                <a
                    :href="feature.link"
                    class="flex items-center justify-center gap-3 bg-[#202020] text-white text-lg px-6 py-2 ml-3 rounded-full w-max self-end justify-self-end"
                >
                    {{ feature.last ? "Discover all features" : "Read More" }}
                    <span class="material-symbols-outlined text-2lg">
                        arrow_forward
                    </span>
                </a>
            </div>
        </div>
    </section>
</template>

<script setup lang="ts">
import { useData } from "vitepress";

const { frontmatter } = useData();
</script>
