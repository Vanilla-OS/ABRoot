<script setup lang="ts">
import { useData } from "vitepress";
import Hero from "./components/Hero.vue";
import Features from "./components/Features.vue";
import CurrentState from "./components/CurrentState.vue";
import Customizable from "./components/Customizable.vue";
import Rebase from "./components/Rebase.vue";
import YourData from "./components/YourData.vue";
import Recover from "./components/Recover.vue";
import ResetRoots from "./components/ResetRoots.vue";
import GetStarted from "./components/GetStarted.vue";
import Help from "./components/Help.vue";

// https://vitepress.dev/reference/runtime-api#usedata
const { site, frontmatter } = useData();
</script>

<template>
    <nav>
        <div
            class="container mx-auto flex items-center justify-between py-4 px-6 flex-wrap"
        >
            <a href="/" class="font-bold text-2xl">
                <img src="/assets/logo.svg" alt="Vib" class="h-14" />
            </a>
            <button
                id="menu-button"
                class="text-gray-300 inline-flex p-3 hover:bg-gray-700 rounded lg:hidden ml-auto hover:text-white outline-none"
            >
                <span class="material-symbols-outlined">menu</span>
            </button>
            <div
                id="menu"
                class="w-full lg:inline-flex lg:flex-grow lg:w-auto hidden"
            >
                <div
                    class="lg:inline-flex lg:flex-row lg:ml-auto lg:w-auto w-full lg:items-center items-start flex flex-col lg:h-auto"
                >
                    <a
                        href="/features"
                        class="text-gray-300 hover:text-white px-3 py-2 rounded"
                        >Features</a
                    >
                    <a
                        href="/examples"
                        class="text-gray-300 hover:text-white px-3 py-2 rounded"
                        >Examples</a
                    >
                    <a
                        href="https://docs.vanillaos.org/docs/en/ABRoot"
                        class="text-gray-300 hover:text-white px-3 py-2 rounded"
                        >Documentation</a
                    >
                    <a
                        href="https://github.com/vanilla-os/ABRoot"
                        class="text-gray-300 hover:text-white px-3 py-2 rounded"
                        >GitHub</a
                    >
                    <a
                        href="#"
                        class="bg-rose-600 hover:bg-rose-500 text-white font-bold px-6 py-2 ml-3 rounded-full transition duration-300"
                        >Get Started</a
                    >
                </div>
            </div>
        </div>
    </nav>
    <div class="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8">
        <div v-if="frontmatter.home">
            <Hero />
            <div class="flex flex-col gap-32 mb-10">
                <Features />
                <CurrentState />
                <Customizable />
                <Rebase />
                <YourData />
                <Recover />
                <ResetRoots />
                <GetStarted />
                <Help />
            </div>
        </div>
        <div v-else>
            <a href="/">Home</a>
            <Content />
        </div>
    </div>
</template>
