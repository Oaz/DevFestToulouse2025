import getpass
import os

from langchain_anthropic import ChatAnthropic
from langchain_core.language_models import BaseChatModel
from Models.Model import Model


class Anthropic(Model):
  def __init__(self, model: str):
    super().__init__(model)

  def create(self, temperature) -> BaseChatModel:
    if not os.environ.get("ANTHROPIC_API_KEY"):
      os.environ["ANTHROPIC_API_KEY"] = getpass.getpass("Enter your Anthropic AI API key: ")
    return ChatAnthropic(
      model_name=self.name,
      temperature=temperature,
    )
