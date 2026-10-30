package com.bookingnetwork.appmarket.feature.detail

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

data class AppDetailState(
    val appId: String,
    val title: String,
    val packageName: String,
    val developerName: String,
    val versionName: String,
    val versionCode: Long,
    val rating: Double,
    val description: String,
    val canDirectInstall: Boolean,
    val isCompatible: Boolean,
    val incompatibleReason: String? = null,
    val sha256Checksum: String,
    val sizeBytes: Long
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AppDetailScreen(
    app: AppDetailState,
    onInstallClick: () -> Unit,
    onNavigateBack: () -> Unit
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(app.title) },
                navigationIcon = {
                    IconButton(onClick = onNavigateBack) {
                        Text("←", style = MaterialTheme.typography.titleLarge)
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp)
                .verticalScroll(rememberScrollState()),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Header Info
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(text = "📦", style = MaterialTheme.typography.displayMedium)
                Spacer(modifier = Modifier.width(16.dp))
                Column {
                    Text(text = app.title, style = MaterialTheme.typography.headlineSmall)
                    Text(text = app.developerName, style = MaterialTheme.typography.bodyMedium)
                    Text(text = "v${app.versionName} • ★ ${app.rating}", style = MaterialTheme.typography.labelMedium)
                }
            }

            // Install or Outlink CTA Button
            if (app.isCompatible) {
                Button(
                    onClick = onInstallClick,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text(if (app.canDirectInstall) "Install APK Directly" else "View on Official Store ↗")
                }
            } else {
                OutlinedButton(
                    onClick = {},
                    enabled = false,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text("Incompatible with your device: ${app.incompatibleReason}")
                }
            }

            // Cryptographic Fingerprint Card
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
            ) {
                Column(modifier = Modifier.padding(12.dp)) {
                    Text(text = "SECURITY & INTEGRITY", style = MaterialTheme.typography.labelSmall)
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(text = "SHA-256: ${app.sha256Checksum}", style = MaterialTheme.typography.bodySmall)
                    Text(text = "Package: ${app.packageName}", style = MaterialTheme.typography.bodySmall)
                }
            }

            // Description
            Text(text = "About this application", style = MaterialTheme.typography.titleMedium)
            Text(text = app.description, style = MaterialTheme.typography.bodyMedium)
        }
    }
}
