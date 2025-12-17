plugins {
    kotlin("multiplatform") version "1.9.0"
}

kotlin {
    jvm()
    sourceSets {
        val commonMain by getting
        val commonTest by getting
    }
}

