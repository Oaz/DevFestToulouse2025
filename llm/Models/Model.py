from langchain_core.language_models import BaseChatModel


class Model:
  def __init__(self, name: str):
    self.name = name

  def create(self, temperature) -> BaseChatModel:
    pass


