import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';

@Component({
  selector: 'tag-list-item',
  templateUrl: './tag-list-item.component.html',
  styleUrls: ['./tag-list-item.component.scss']
})
export class TagListItemComponent implements OnInit {

  hovered = false;

  keyFocused = false;
  valueFocused = false;

  keyValue: string;
  valueValue: string;

  @Output()
  keyChange = new EventEmitter<string>();
  @Output()
  valueChange = new EventEmitter<string>();
  @Output()
  deleted = new EventEmitter<void>();

  constructor() { }

  ngOnInit() {
  }

  @Input()
  get key() {
    return this.keyValue;
  }

  set key(val) {
    this.keyValue = val;
    this.keyChange.emit(this.keyValue);
  }

  @Input()
  get value() {
    return this.valueValue;
  }

  set value(val) {
    this.valueValue = val;
    this.valueChange.emit(this.valueValue);
  }

  isKeyEmpty(): boolean {
    return this.key.trim().length === 0;
  }

  isValueEmpty(): boolean {
    return this.value.trim().length === 0;
  }

  showDeleteIcon(): boolean {
    return this.hovered || this.keyFocused || this.valueFocused;
  }

  onDeletePressed(): void {
    this.deleted.emit();
  }
}
