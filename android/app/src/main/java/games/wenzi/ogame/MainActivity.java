package games.wenzi.ogame;

import android.animation.Animator;
import android.animation.AnimatorListenerAdapter;
import android.animation.ObjectAnimator;
import android.graphics.Color;
import android.os.Bundle;
import android.view.View;
import android.view.Window;
import android.view.animation.AccelerateDecelerateInterpolator;
import android.webkit.ValueCallback;
import android.webkit.WebChromeClient;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.net.Uri;
import android.content.Intent;
import androidx.core.splashscreen.SplashScreen;
import androidx.core.view.WindowCompat;
import androidx.core.view.WindowInsetsControllerCompat;
import com.getcapacitor.BridgeActivity;

public class MainActivity extends BridgeActivity {
    private boolean isWebViewReady = false;
    private static final int FILE_CHOOSER_REQUEST_CODE = 1001;
    private ValueCallback<Uri[]> filePathCallback;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        // 安装 SplashScreen，必须在 super.onCreate 之前调用
        SplashScreen splashScreen = SplashScreen.installSplashScreen(this);

        // 保持 SplashScreen 直到 WebView 加载完成
        splashScreen.setKeepOnScreenCondition(() -> !isWebViewReady);

        // 设置淡出退出动画
        splashScreen.setOnExitAnimationListener(splashScreenView -> {
            // 创建淡出动画
            ObjectAnimator fadeOut = ObjectAnimator.ofFloat(
                splashScreenView.getView(),
                View.ALPHA,
                1f,
                0f
            );
            fadeOut.setInterpolator(new AccelerateDecelerateInterpolator());
            fadeOut.setDuration(300);

            // 动画结束后移除 SplashScreen
            fadeOut.addListener(new AnimatorListenerAdapter() {
                @Override
                public void onAnimationEnd(Animator animation) {
                    splashScreenView.remove();
                }
            });

            fadeOut.start();
        });

        super.onCreate(savedInstanceState);

        Window window = getWindow();

        // 启用边到边显示（Edge-to-Edge）
        WindowCompat.setDecorFitsSystemWindows(window, false);

        // 设置透明状态栏和导航栏
        window.setStatusBarColor(Color.TRANSPARENT);
        window.setNavigationBarColor(Color.TRANSPARENT);

        // 设置状态栏图标为浅色（因为背景是深色）
        WindowInsetsControllerCompat controller = WindowCompat.getInsetsController(window, window.getDecorView());
        if (controller != null) {
            controller.setAppearanceLightStatusBars(false);
            controller.setAppearanceLightNavigationBars(false);
        }
    }

    @Override
    public void onStart() {
        super.onStart();
        WebView webView = getBridge().getWebView();
        if (webView != null) {
            WebSettings settings = webView.getSettings();
            // 禁用 WebView 文本缩放，防止系统字体大小设置影响布局
            settings.setTextZoom(100);
            // 优化 WebView 性能
            settings.setCacheMode(WebSettings.LOAD_DEFAULT);
            settings.setDomStorageEnabled(true);
            settings.setDatabaseEnabled(true);
            // 启用硬件加速渲染
            webView.setLayerType(android.view.View.LAYER_TYPE_HARDWARE, null);

            // 监听页面加载进度，加载完成后隐藏 SplashScreen
            webView.setWebChromeClient(new WebChromeClient() {
                @Override
                public void onProgressChanged(WebView view, int newProgress) {
                    super.onProgressChanged(view, newProgress);
                    // 当页面加载达到 80% 时认为可以显示
                    if (newProgress >= 80) {
                        isWebViewReady = true;
                    }
                }

                @Override
                public boolean onShowFileChooser(WebView webView, ValueCallback<Uri[]> filePathCallback,
                                                 FileChooserParams fileChooserParams) {
                    // 清理之前的回调
                    if (MainActivity.this.filePathCallback != null) {
                        MainActivity.this.filePathCallback.onReceiveValue(null);
                    }

                    MainActivity.this.filePathCallback = filePathCallback;

                    // 创建文件选择器 Intent
                    Intent intent = fileChooserParams.createIntent();
                    try {
                        startActivityForResult(intent, FILE_CHOOSER_REQUEST_CODE);
                    } catch (android.content.ActivityNotFoundException e) {
                        MainActivity.this.filePathCallback = null;
                        return false;
                    }
                    return true;
                }
            });
        }
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);

        if (requestCode == FILE_CHOOSER_REQUEST_CODE) {
            if (filePathCallback == null) {
                return;
            }

            Uri[] results = null;
            if (resultCode == RESULT_OK && data != null) {
                String dataString = data.getDataString();
                if (dataString != null) {
                    results = new Uri[]{Uri.parse(dataString)};
                }
            }

            filePathCallback.onReceiveValue(results);
            filePathCallback = null;
        }
    }
}
