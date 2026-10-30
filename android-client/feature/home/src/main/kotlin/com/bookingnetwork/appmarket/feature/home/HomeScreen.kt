package com.bookingnetwork.appmarket.feature.home

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

data class AppFeedItem(
    val appId: String,
    val slug: String,
    val title: String,
    val category: String,
    val rating: Double,
    val isDirectDownload: Boolean
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    featuredApps: List<AppFeedItem>,
    topCharts: List<AppFeedItem>,
    onAppClick: (String) -> Unit
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("BN App Market") },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.primaryContainer,
                    titleContentColor = MaterialTheme.colorScheme.onPrimaryContainer
                )
            )
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(24.dp)
        ) {
            item {
                Text(
                    text = "Featured & Editors Choice",
                    style = MaterialTheme.typography.titleLarge,
                    modifier = Modifier.padding(top = 16.dp)
                )
                Spacer(modifier = Modifier.height(12.dp))
                LazyRow(
                    horizontalArrangement = Arrangement.spacedBy(16.dp)
                ) {
                    items(featuredApps) { app ->
                        FeaturedCard(app = app, onClick = { onAppClick(app.slug) })
                    }
                }
            }

            item {
                Text(
                    text = "Top Charts",
                    style = MaterialTheme.typography.titleLarge
                )
            }

            items(topCharts) { app ->
                AppListItem(app = app, onClick = { onAppClick(app.slug) })
            }
        }
    }
}

@Composable
fun FeaturedCard(app: AppFeedItem, onClick: () -> Unit) {
    Card(
        modifier = Modifier
            .width(180.dp)
            .clickable(onClick = onClick),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(text = "📦", style = MaterialTheme.typography.headlineMedium)
            Spacer(modifier = Modifier.height(8.dp))
            Text(text = app.title, style = MaterialTheme.typography.titleMedium, maxLines = 1)
            Text(text = app.category, style = MaterialTheme.typography.bodySmall)
            Spacer(modifier = Modifier.height(4.dp))
            Text(text = "★ ${app.rating}", style = MaterialTheme.typography.labelSmall)
        }
    }
}

@Composable
fun AppListItem(app: AppFeedItem, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(text = "📱", style = MaterialTheme.typography.headlineSmall)
        Spacer(modifier = Modifier.width(16.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(text = app.title, style = MaterialTheme.typography.titleMedium)
            Text(text = "${app.category} • ★ ${app.rating}", style = MaterialTheme.typography.bodySmall)
        }
        AssistChip(
            onClick = onClick,
            label = { Text(if (app.isDirectDownload) "Install" else "View ↗") }
        )
    }
}
