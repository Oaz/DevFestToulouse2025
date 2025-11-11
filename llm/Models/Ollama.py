from langchain_core.language_models import BaseChatModel
from langchain_ollama import ChatOllama
from Models.Model import Model


class Ollama(Model):
  def __init__(self, model: str):
    super().__init__(model)

  def create(self, temperature) -> BaseChatModel:
    return ChatOllama(
      model=self.name,
      temperature=temperature,
      top_p=0.9,
      top_k=40,
    )
