package com.bookingnetwork.appmarket.feature.updates

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

data class AppUpdateItem(
    val packageName: String,
    val title: String,
    val currentVersion: String,
    val latestVersion: String,
    val hasSignerConflict: Boolean,
    val sizeBytes: Long
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun UpdatesCenterScreen(
    compatibleUpdates: List<AppUpdateItem>,
    conflictingUpdates: List<AppUpdateItem>,
    onUpdateClick: (AppUpdateItem) -> Unit,
    onUpdateAllClick: () -> Unit
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("App Updates") }
            )
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(20.dp)
        ) {
            // Section 1: Direct 1-Click Compatible Updates
            item {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = "Updates Available (${compatibleUpdates.size})",
                        style = MaterialTheme.typography.titleMedium
                    )
                    if (compatibleUpdates.isNotEmpty()) {
                        Button(onClick = onUpdateAllClick) {
                            Text("Update All")
                        }
                    }
                }
            }

            items(compatibleUpdates) { app ->
                UpdateCard(app = app, onUpdateClick = { onUpdateClick(app) })
            }

            // Section 2: Signature Conflicts (Installed from Google Play or other stores)
            if (conflictingUpdates.isNotEmpty()) {
                item {
                    Text(
                        text = "Different Distributor Signature (${conflictingUpdates.size})",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.error,
                        modifier = Modifier.padding(top = 16.dp)
                    )
                    Text(
                        text = "These apps were installed from a different store with an incompatible signature key. Data backup and reinstallation required.",
                        style = MaterialTheme.typography.bodySmall,
                        modifier = Modifier.padding(top = 4.dp)
                    )
                }

                items(conflictingUpdates) { app ->
                    ConflictingUpdateCard(app = app)
                }
            }
        }
    }
}

@Composable
fun UpdateCard(app: AppUpdateItem, onUpdateClick: () -> Unit) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(text = app.title, style = MaterialTheme.typography.titleSmall)
                Text(
                    text = "v${app.currentVersion} → v${app.latestVersion}",
                    style = MaterialTheme.typography.bodySmall
                )
            }
            Button(onClick = onUpdateClick) {
                Text("Update")
            }
        }
    }
}

@Composable
fun ConflictingUpdateCard(app: AppUpdateItem) {
    OutlinedCard(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.outlinedCardColors(
            containerColor = MaterialTheme.colorScheme.errorContainer.copy(alpha = 0.2f)
        )
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(text = app.title, style = MaterialTheme.typography.titleSmall)
            Text(
                text = "Installed: v${app.currentVersion} | Available: v${app.latestVersion}",
                style = MaterialTheme.typography.bodySmall
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "⚠️ Signer conflict: Android blocks in-place update. Uninstall existing copy to switch distribution sources.",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.error
            )
        }
    }
}

