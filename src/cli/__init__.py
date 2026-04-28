"""量潮数字资产云 CLI"""

__version__ = "0.1.0"
__stage__ = "alpha"
__stage_description__ = "本机验证"

def get_app():
    from .app.cli import app
    return app

__all__ = ["app", "__version__", "__stage__", "__stage_description__", "get_app"]
