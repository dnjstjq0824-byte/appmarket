pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}
dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "BNAppMarket"
include(":app")
include(":core:device")
include(":core:downloader")
include(":core:installer")
include(":feature:home")
include(":feature:detail")
include(":feature:updates")
