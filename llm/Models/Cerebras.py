import getpass
import os

from langchain_cerebras import ChatCerebras
from langchain_core.language_models import BaseChatModel
from Models.Model import Model


class Cerebras(Model):
  def __init__(self, model: str):
    super().__init__(model)

  def create(self, temperature) -> BaseChatModel:
    if not os.environ.get("CEREBRAS_API_KEY"):
      os.environ["CEREBRAS_API_KEY"] = getpass.getpass("Enter your Cerebras API key: ")
    return ChatCerebras(
      model=self.name,
      temperature=temperature,
    )
