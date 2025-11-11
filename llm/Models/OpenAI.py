import getpass
import os
from langchain_core.language_models import BaseChatModel
from langchain_openai import ChatOpenAI
from Models.Model import Model


class OpenAI(Model):
  def __init__(self, model: str):
    super().__init__(model)

  def create(self, temperature) -> BaseChatModel:
    if not os.environ.get("OPENAI_API_KEY"):
      os.environ["OPENAI_API_KEY"] = getpass.getpass("Enter your OpenAI API key: ")
    return ChatOpenAI(
      model=self.name,
      temperature=temperature,
      top_p=1.0,
      max_tokens=None,
      presence_penalty=0.0,
      frequency_penalty=0.0
    )
