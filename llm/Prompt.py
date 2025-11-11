from typing import Callable, List
from data_model import State


def remove_thinking(text: str) -> str:
  while '<thinking>' in text and '</thinking>' in text:
    start = text.find('<thinking>')
    end = text.find('</thinking>') + len('</thinking>')
    text = text[:start] + text[end:]
  return text


class PromptSection:
  def __init__(self, name: str, content: Callable[[State], str], condition: Callable[[State], bool]):
    self.name = name.upper()
    self.content = content
    self.condition = condition


class Prompt:
  def __init__(self):
    self.sections : List[PromptSection] = []

  def add_raw_section(self, section: PromptSection):
    self.sections.append(section)

  def add_section(self, name: str, condition: Callable[[State], bool], content: Callable[[State], str]):
    self.add_raw_section(PromptSection(name=name, content=content, condition=condition))

  def add_conditional_section(self, name: str, condition: Callable[[State], bool], content: str):
    self.add_raw_section(PromptSection(name=name, content=lambda _: content, condition=condition))

  def add_dynamic_section(self, name: str, content: Callable[[State], str]):
    self.add_raw_section(PromptSection(name=name, content=content, condition=lambda _: True))

  def add_static_section(self, name: str, content: str):
    self.add_raw_section(PromptSection(name=name, content=lambda _: content, condition=lambda _: True))

  def format(self, state: State):
    result = ""
    for section in self.sections:
      if section.condition(state):
        result += f"<{section.name}>\n{section.content(state)}\n</{section.name}>\n"
    return result

