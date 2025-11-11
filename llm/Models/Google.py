import getpass
import os
from langchain_core.language_models import BaseChatModel
from langchain_google_genai import ChatGoogleGenerativeAI
from Models.Model import Model


class Google(Model):
  def __init__(self, model: str):
    super().__init__(model)

  def create(self, temperature) -> BaseChatModel:
    if not os.environ.get("GOOGLE_API_KEY"):
      os.environ["GOOGLE_API_KEY"] = getpass.getpass("Enter your Google AI API key: ")
    return ChatGoogleGenerativeAI(
      model=self.name,
      temperature=temperature,
      max_tokens=None,
    )
